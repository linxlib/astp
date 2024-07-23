package astp

// Receiver 接收器
type Receiver struct {
	Name       string  //参数名
	Pointer    bool    //是否指针
	TypeString string  //类型名
	Type       *Struct `json:"-"` //类型
}
