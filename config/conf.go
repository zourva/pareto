package config

type DataType = string

const (
	Json    DataType = "json"
	Yaml    DataType = "yaml" // for both yaml and yml
	MsgPack DataType = "msgpack"
)

type FileType = string

const (
	Text   FileType = "text"
	Boltdb FileType = "boltdb"
	Sqlite FileType = "sqlite"
)
