package config

type DBCodec struct{}

func (DBCodec) Encode(v map[string]any) ([]byte, error) {
	//return yaml.Marshal(v)
	panic("not supported")
	return nil, nil
}

func (DBCodec) Decode(b []byte, v map[string]any) error {
	//return yaml.Unmarshal(b, &v)
	panic("not supported")
	return nil
}
