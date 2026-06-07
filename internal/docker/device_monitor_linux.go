//go:build linux

package docker

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unsafe"

	componenttree "helios/internal/component_tree"
	"helios/internal/logger"

	"golang.org/x/sys/unix"
)

// DeviceMonitor watches /dev (and configured device subdirectories) via inotify
// and creates device nodes inside containers when devices appear or change.
type DeviceMonitor struct {
	ctx       context.Context
	cancel    context.CancelFunc
	inoFd     int
	epFd      int
	client    *DockerClient
	tree      *componenttree.ComponentTree
	watchDirs map[int32]string // inotify watch descriptor → directory path
}

func newDeviceMonitor(ctx context.Context, client *DockerClient, tree *componenttree.ComponentTree) *DeviceMonitor {
	ctx, cancel := context.WithCancel(ctx)
	return &DeviceMonitor{
		ctx:    ctx,
		cancel: cancel,
		inoFd:  -1,
		epFd:   -1,
		client: client,
		tree:   tree,
	}
}

func (m *DeviceMonitor) start() error {
	inoFd, err := unix.InotifyInit1(unix.IN_CLOEXEC | unix.IN_NONBLOCK)
	if err != nil {
		return fmt.Errorf("inotify_init1: %w", err)
	}

	devWd, err := unix.InotifyAddWatch(inoFd, "/dev", unix.IN_CREATE|unix.IN_MOVED_TO)
	if err != nil {
		unix.Close(inoFd)
		return fmt.Errorf("inotify_add_watch /dev: %w", err)
	}

	watchDirs := map[int32]string{int32(devWd): "/dev"}

	// Also watch any configured device source that is a directory (e.g. /dev/snd).
	_ = m.tree.EachComponent(func(_ string, _ string, comp *componenttree.Component) error {
		spec := comp.GetDockerSpec()
		if spec == nil {
			return nil
		}
		for _, device := range spec.Devices {
			info, err := os.Stat(device.Source)
			if err != nil || !info.IsDir() {
				continue
			}
			wd, err := unix.InotifyAddWatch(inoFd, device.Source, unix.IN_CREATE|unix.IN_MOVED_TO)
			if err != nil {
				logger.Warnw("inotify_add_watch failed for device directory", "path", device.Source, "error", err)
				continue
			}
			watchDirs[int32(wd)] = device.Source
		}
		return nil
	})

	epFd, err := unix.EpollCreate1(unix.EPOLL_CLOEXEC)
	if err != nil {
		unix.Close(inoFd)
		return fmt.Errorf("epoll_create1: %w", err)
	}

	ev := unix.EpollEvent{Events: unix.EPOLLIN, Fd: int32(inoFd)}
	if err := unix.EpollCtl(epFd, unix.EPOLL_CTL_ADD, inoFd, &ev); err != nil {
		unix.Close(inoFd)
		unix.Close(epFd)
		return fmt.Errorf("epoll_ctl: %w", err)
	}

	m.inoFd = inoFd
	m.epFd = epFd
	m.watchDirs = watchDirs

	go m.run()
	logger.Info("Device monitor started, watching /dev")
	return nil
}

func (m *DeviceMonitor) stop() {
	m.cancel()
	if m.epFd >= 0 {
		unix.Close(m.epFd)
		m.epFd = -1
	}
	if m.inoFd >= 0 {
		unix.Close(m.inoFd)
		m.inoFd = -1
	}
}

func (m *DeviceMonitor) run() {
	buf := make([]byte, 4096)
	epEvents := make([]unix.EpollEvent, 8)

	for {
		select {
		case <-m.ctx.Done():
			return
		default:
		}

		// Wait up to 500 ms so we can re-check context without burning CPU.
		n, err := unix.EpollWait(m.epFd, epEvents, 500)
		if err != nil {
			if err == unix.EINTR {
				continue
			}
			return
		}
		if n == 0 {
			continue
		}

		nread, err := unix.Read(m.inoFd, buf)
		if err != nil {
			if err == unix.EINTR || err == unix.EAGAIN {
				continue
			}
			return
		}

		m.parseEvents(buf[:nread])
	}
}

func (m *DeviceMonitor) parseEvents(buf []byte) {
	offset := 0
	for offset+unix.SizeofInotifyEvent <= len(buf) {
		event := (*unix.InotifyEvent)(unsafe.Pointer(&buf[offset]))
		nameLen := int(event.Len)
		name := ""
		if nameLen > 0 && offset+unix.SizeofInotifyEvent+nameLen <= len(buf) {
			raw := buf[offset+unix.SizeofInotifyEvent : offset+unix.SizeofInotifyEvent+nameLen]
			if i := bytes.IndexByte(raw, 0); i >= 0 {
				name = string(raw[:i])
			} else {
				name = string(raw)
			}
		}
		offset += unix.SizeofInotifyEvent + nameLen

		if name != "" {
			baseDir := m.watchDirs[event.Wd]
			if baseDir == "" {
				baseDir = "/dev"
			}
			m.handleDeviceAppeared(filepath.Join(baseDir, name))
		}
	}
}

func (m *DeviceMonitor) handleDeviceAppeared(hostPath string) {
	major, minor, err := resolveDeviceMajorMinor(hostPath)
	if err != nil {
		// Not a device file (directory, non-device symlink, etc.).
		return
	}

	_ = m.tree.EachComponent(func(_ string, name string, comp *componenttree.Component) error {
		spec := comp.GetDockerSpec()
		if spec == nil {
			return nil
		}
		containerID := comp.GetDockerConnID()
		if containerID == "" {
			return nil
		}
		for _, device := range spec.Devices {
			var targetPath string
			if deviceMatchesHost(device.Source, hostPath) {
				// File device: source directly matches the new path.
				targetPath = device.Target
			} else if isSubpath(device.Source, hostPath) {
				// Directory device (e.g. /dev/snd): map the relative path into the target dir.
				rel, _ := filepath.Rel(device.Source, hostPath)
				targetPath = filepath.Join(device.Target, rel)
			}
			if targetPath == "" {
				continue
			}
			logger.Infow("Hot-plug: device appeared, creating node in container",
				"device", hostPath, "container", name, "target", targetPath)
			if err := m.client.execMknod(containerID, targetPath, major, minor); err != nil {
				logger.Warnw("Failed to create hot-plug device node",
					"error", err, "container", name, "device", targetPath)
			}
		}
		return nil
	})
}

// deviceMatchesHost reports whether the configured source path refers to hostPath,
// accounting for the fact that source may be a symlink that resolves to hostPath.
func deviceMatchesHost(source, hostPath string) bool {
	if source == hostPath {
		return true
	}
	if resolved, err := filepath.EvalSymlinks(source); err == nil && resolved == hostPath {
		return true
	}
	if resolved, err := filepath.EvalSymlinks(hostPath); err == nil && resolved == source {
		return true
	}
	return false
}

// isSubpath reports whether child is directly inside parent (not parent itself).
func isSubpath(parent, child string) bool {
	rel, err := filepath.Rel(parent, child)
	if err != nil {
		return false
	}
	return rel != "." && !strings.HasPrefix(rel, "..")
}
