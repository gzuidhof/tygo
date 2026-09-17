package unionpackage

//tygo:union
type Message interface {
	isMessage()
}
