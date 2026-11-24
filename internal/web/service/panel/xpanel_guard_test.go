package panel
import "testing"
func TestXpanelUpdateGuard(t *testing.T){s:=PanelService{};for _,dev:=range []bool{false,true}{id,err:=s.StartUpdateChannel(dev);if id!=0||err==nil{t.Fatal("branded updater must refuse before network or process effects")}}}
