package general

// Alloc ...
type Alloc struct {
	Balance string `json:"balance"`
}

// Genesis ...
type Genesis struct {
	Config struct {
		ChainID             int `json:"chainId"`
		HomesteadBlock      int `json:"homesteadBlock"`
		Eip150Block         int `json:"eip150Block"`
		Eip155Block         int `json:"eip155Block"`
		Eip158Block         int `json:"eip158Block"`
		ByzantiumBlock      int `json:"byzantiumBlock"`
		ConstantinopleBlock int `json:"constantinopleBlock"`
		PetersburgBlock     int `json:"petersburgBlock"`
		Ethash              struct {
		} `json:"ethash"`
	} `json:"config"`
	Difficulty string            `json:"difficulty"`
	GasLimit   string            `json:"gasLimit"`
	Alloc      map[string]*Alloc `json:"alloc"`
}

// LoadGenesis ...
func LoadGenesis() {

}
