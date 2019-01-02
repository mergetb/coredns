package nex

/* ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~*\
 * Network Object Interface
\* ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~*/

type NetworkObj struct {
	*Network
	Version int64
}

func NewNetworkObj(n *Network) *NetworkObj { return &NetworkObj{n, 0} }
func (n *NetworkObj) Key() string          { return "/net/" + n.Name }
func (n *NetworkObj) GetVersion() int64    { return n.Version }
func (n *NetworkObj) SetVersion(v int64)   { n.Version = v }
func (n *NetworkObj) Value() interface{}   { return n.Network }

/* ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~*\
 * Member Object Interfaces
\* ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~*/

type MemberObj struct {
	*Member
	Version int64
}

func (m *MemberObj) GetVersion() int64  { return m.Version }
func (m *MemberObj) SetVersion(v int64) { m.Version = v }

// MacIndex stores and retrieves network members based on their MAC address.
// MAC addresses are assumed to be global per Nex instance.
type MacIndex struct{ MemberObj }

func NewMacIndex(m *Member) *MacIndex  { return &MacIndex{MemberObj{m, 0}} }
func (m *MacIndex) Key() string        { return "/member/" + m.Mac }
func (m *MacIndex) Value() interface{} { return m.Member }

// Ip4Index stores and retrieves network members based on their IPv4 address.
// Ipv4 addresses are assumed to be global mer Nex Instance
type Ip4Index struct{ MemberObj }

func NewIp4Index(m *Member) *Ip4Index { return &Ip4Index{MemberObj{m, 0}} }
func (m *Ip4Index) Key() string {
	if m.Ip4 != nil {
		return "/member/ip4/" + m.Ip4.Address
	}
	return ""
}
func (m *Ip4Index) Value() interface{} { return m.Mac }

// NameIndex stores and retrieves network members based on their dns name. DNS
// names are assumed to be global per Nex Instance.
type NameIndex struct{ MemberObj }

func NewNameIndex(m *Member) *NameIndex { return &NameIndex{MemberObj{m, 0}} }
func (m *NameIndex) Key() string        { return "/member/name/" + m.Name }
func (m *NameIndex) Value() interface{} { return m.Mac }

// NetIndex stores and retrieves network members based on their parent network.
type NetIndex struct{ MemberObj }

func NewNetIndex(m *Member) *NetIndex  { return &NetIndex{MemberObj{m, 0}} }
func (m *NetIndex) Key() string        { return "/member/net/" + m.Net + "/" + m.Mac }
func (m *NetIndex) Value() interface{} { return m.Mac }

func IndexExists(o Object) bool {
	return o.Key() != ""
}

/* ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~*\
 * Pool Object Interface
\* ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~*/

type PoolObj struct {
	*Pool
	Version int64
}

func NewPoolObj(p *Pool) *PoolObj     { return &PoolObj{p, 0} }
func (p *PoolObj) GetVersion() int64  { return p.Version }
func (p *PoolObj) SetVersion(v int64) { p.Version = v }
func (p *PoolObj) Key() string        { return "/pool/" + p.Net }
func (p *PoolObj) Value() interface{} { return p.Pool }
