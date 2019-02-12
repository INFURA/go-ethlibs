package eth

type Address Data20

func (a Address) String() string {
	return string(a)
}

func (a *Address) UnmarshalJSON(data []byte) error {
	str, err := unmarshalHex(data, 20, "data")
	if err != nil {
		return err
	}
	*a = Address(str)
	return nil
}
