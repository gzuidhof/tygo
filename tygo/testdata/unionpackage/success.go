package unionpackage

type Success struct {
	Result string `json:"result"`
}

func (Success) isMessage() {}
