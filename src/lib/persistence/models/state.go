package models

type State string

const (
	AC State = "AC"
	AL State = "AL"
	AP State = "AP"
	AM State = "AM"
	BA State = "BA"
	CE State = "CE"
	DF State = "DF"
	ES State = "ES"
	GO State = "GO"
	MA State = "MA"
	MT State = "MT"
	MS State = "MS"
	MG State = "MG"
	PA State = "PA"
	PB State = "PB"
	PR State = "PR"
	PE State = "PE"
	PI State = "PI"
	RJ State = "RJ"
	RN State = "RN"
	RS State = "RS"
	RO State = "RO"
	RR State = "RR"
	SC State = "SC"
	SP State = "SP"
	SE State = "SE"
	TO State = "TO"
)

// GetStateList ...
func GetStateList() []State {
	return []State{AC, AL, AP, AM, BA, CE, DF, ES, GO, MA, MT, MS, MG, PA, PB, PR, PE, PI, RJ, RN, RS, RO, RR, SC, SP, SE, TO}
}

// GetState ...
func GetState(s string) (State, bool) {
	states := GetStateList()
	for _, state := range states {
		if string(state) == s {
			return state, true
		}
	}
	return "", false
}

// Name returns the name of the State
func (s *State) Name() string {
	return GetStateName(*s)
}

// GetStateName returns the name of the given State
func GetStateName(state State) string {
	switch state {
	case AC:
		return "Acre"
	case AL:
		return "Alagoas"
	case AP:
		return "Amapá"
	case AM:
		return "Amazonas"
	case BA:
		return "Bahia"
	case CE:
		return "Ceará"
	case DF:
		return "Distrito Federal"
	case ES:
		return "Espírito do Santo"
	case GO:
		return "Goiás"
	case MA:
		return "Maranhão"
	case MT:
		return "Mato Grosso"
	case MS:
		return "Mato Grosso do Sul"
	case MG:
		return "Minas Gerais"
	case PA:
		return "Pará"
	case PB:
		return "Paraíba"
	case PR:
		return "Paraná"
	case PE:
		return "Pernambuco"
	case PI:
		return "Piauí"
	case RJ:
		return "Rio de Janeiro"
	case RN:
		return "Rio Grande do Norte"
	case RS:
		return "Rio Grade do Sul"
	case RO:
		return "Rondônia"
	case RR:
		return "Roraima"
	case SC:
		return "Santa Catarina"
	case SP:
		return "São Paulo"
	case SE:
		return "Sergipe"
	case TO:
		return "Tocantins"
	}
	return ""
}
