package service

// Descriptor defines meta description info
// needed to define a service.
type Descriptor struct {
	Name     string `json:"name"`
	Registry string `json:"registry"`
}
