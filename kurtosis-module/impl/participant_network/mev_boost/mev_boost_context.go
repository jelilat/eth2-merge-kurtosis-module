package mev_boost

// Still needs to be checked that it's collecting the right info
type MEVBoostContext struct {
	// Information about the mev-boost node that the CL client needs, analogous to ELClientContext
	clientName       string
	enr              string // Ethereum Node Records
	enode            string
	ipAddr           string
	rpcPortNum       uint16
	wsPortNum        uint16
	engineRpcPortNum uint16
}

func NewMEVBoostContext(clientName string, enr string, enode string, ipAddr string, rpcPortNum uint16, wsPortNum uint16, engineRpcPortNum uint16) *MEVBoostContext {
	return &MEVBoostContext{clientName: clientName, enr: enr, enode: enode, ipAddr: ipAddr, rpcPortNum: rpcPortNum, wsPortNum: wsPortNum, engineRpcPortNum: engineRpcPortNum}
}

func (ctx *MEVBoostContext) GetClientName() string {
	return ctx.clientName
}

func (ctx *MEVBoostContext) GetENR() string {
	return ctx.enr
}

func (ctx *MEVBoostContext) GetEnode() string {
	return ctx.enode
}

func (ctx *MEVBoostContext) GetIPAddress() string {
	return ctx.ipAddr
}

func (ctx *MEVBoostContext) GetRPCPortNum() uint16 {
	return ctx.rpcPortNum
}

func (ctx *MEVBoostContext) GetWSPortNum() uint16 {
	return ctx.wsPortNum
}

func (ctx *MEVBoostContext) GetEngineRPCPortNum() uint16 {
	return ctx.engineRpcPortNum
}
