package service

func (s *JMService) GetEncryptedConfigValue(encryptKey, encryptedValue string) (resp ResultValue, err error) {
	data := map[string]string{
		"secret_encrypt_key": encryptKey,
		"encrypted_value":    encryptedValue,
	}
	// "/api/v1/terminal/encrypted-config/"
	_, err = s.authClient.Post(TerminalEncryptedConfigURL, data, &resp)
	return
}

type ResultValue struct {
	Value string `json:"value"`
}
