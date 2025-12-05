package client

import (

)

type LocalInterface struct {
	runtimeHash string
}

func initializeLocalClient(hash string) *LocalInterface {
	return &LocalInterface{
		runtimeHash: hash,
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