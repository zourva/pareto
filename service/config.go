package service

// Descriptor defines meta description info
// needed to define a service.
type Descriptor struct {
	Name     string `json:"name"`     // 服务名称
	Domain   int    `json:"domain"`   // 服务归属域
	Registry string `json:"registry"` //
}
