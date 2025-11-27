package client

import (

)

type LocalInterface struct {
	runtime_hash string
}

func initializeLocalClient(runtime_hash string) *LocalInterface {
	return &LocalInterface{
		runtime_hash: runtime_hash,
	}
}

func (x *LocalInterface) Initialize() {
	return
}

func (x *LocalInterface) InitializeComponentTree(path string) {
	return
}

func (x *LocalInterface) StartComponent(name string) {
	return
}

func (x *LocalInterface) StartAllComponents() {
	return
}

func (x *LocalInterface) StopComponent(name string) {
	return
}

func (x *LocalInterface) KillComponent(name string) {
	return
}

func (x *LocalInterface) Clean() {
	return
}

func (x *LocalInterface) Close() {

}