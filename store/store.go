package store

import (
	"strings"
)

type Store interface {
	SaveFile(file *strings.Reader) error
	//storePaht:对象存储的路劲，loadPath:本地路径
	LoadFile(storePath, loadPath string) error

	LoadFolder(storePath, loadPath string) error
}
