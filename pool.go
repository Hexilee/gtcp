package gtcp

import "context"


func InitPool(size uint) {
	InitConnPool(size)
	InitCtrlPool(size)
}

func InitPoolWithCtx(size uint, ctx context.Context) {
	InitConnPoolWithCtx(size, ctx)
	InitCtrlPoolWithCtx(size, ctx)
}

func OpenPool() {
	OpenConnPool()
	OpenCtrlPool()
}

func ClosePool() {
	CloseCtrlPool()
	CloseConnPool()
}

func ReopenPool() {
	ClosePool()
	OpenConnPool()
	OpenCtrlPool()
}

func ReopenPoolWithCtx(ctx context.Context) {
	ClosePool()
	ctrlP.InstallCtx(ctx)
	connP.InstallCtx(ctx)
	OpenConnPool()
	OpenCtrlPool()
}

func DropPool() {
	DropCtrlPool()
	DropConnPool()
}

func ReInitPool(size uint) {
	DropPool()
	InitConnPool(size)
	InitCtrlPool(size)
}

func ReInitPoolWithCtx(size uint, ctx context.Context) {
	DropPool()
	InitConnPoolWithCtx(size, ctx)
	InitCtrlPoolWithCtx(size, ctx)
}
