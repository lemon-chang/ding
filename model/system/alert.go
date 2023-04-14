package system

type Check interface {
	Check(p ApiStatInfo) //放了一个函数
}

type Alert struct {
	AlertHandlers []AlertHandler
}

func (a *Alert) addAlterHandler(alertHandler AlertHandler) {
	a.AlertHandlers = append(a.AlertHandlers, alertHandler)
}
func (a *Alert) Check(p ApiStatInfo) {
	for _, handler := range a.AlertHandlers {
		handler.Check(p)
	}
}

//相当于是一个抽象类
type AlertHandler struct {
	rule         AlertRule
	notification Notification
}

func (aH *AlertHandler) Check(p ApiStatInfo) {

}

//具体的类实现抽象类
//完善信息提醒
type PerfectInformationHandler struct {
	AlertHandler //我们使用组合来达到继承的效果
}

func (aH *PerfectInformationHandler) Check(apiStatInfo ApiStatInfo) {
	if apiStatInfo.unPerfectInformationCount > 0 {
		//发送通知
		aH.notification.notify()
	}

}

type ApiStatInfo struct {
	errCount                  int
	unPerfectInformationCount int //未完善信息数量
}
type AlertRule struct {
}
type Notification struct {
}

func (n *Notification) notify() {

}
