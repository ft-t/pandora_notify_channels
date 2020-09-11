package db

type TgChat struct {
	Id int
}

type PandoraAccount struct {
	Id       int
	Login    string
	Password string
}

type PandoraDevice struct {
	Id               int64 // pandora id
	Name             string
	IsActive         bool
	PandoraAccountId int
}

type PandoraNotification struct {
	Id        int
	Name      string
	DeviceId  int64
	TgChatId  int
	EventType int // todo
}
