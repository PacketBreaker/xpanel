package web

import (
 "net/http/httptest"
 "net/http"
 "io"
 "fmt"
 "strings"
 "net"
 "encoding/json"
 "strconv"
 "github.com/mhsanaei/3x-ui/v3/internal/web/service"
 "os"
 "path/filepath"
 "testing"
 "time"
 "sync/atomic"

 "github.com/mhsanaei/3x-ui/v3/internal/database"
 "github.com/mhsanaei/3x-ui/v3/internal/database/model"
 "github.com/mhsanaei/3x-ui/v3/internal/util/crypto"
 "github.com/robfig/cron/v3"
 "github.com/mhsanaei/3x-ui/v3/internal/web/global"
 "github.com/mhsanaei/3x-ui/v3/internal/web/runtime"
)

// Development-only visual harness. Production router, no scheduler or data-plane startup.
func TestXpanelVisualLocalHarness(t *testing.T) {
 out:=os.Getenv("XPANEL_VISUAL_URL_FILE"); if out=="" {t.Skip("opt-in owned-loopback visual harness")}
 dir:=t.TempDir();t.Setenv("XUI_DB_FOLDER",dir)
 if err:=database.InitDB(filepath.Join(dir,"x-ui.db"));err!=nil{t.Fatal(err)}
 defer database.CloseDB()
 hash,err:=crypto.HashPasswordAsBcrypt("synthetic-local-test-password");if err!=nil{t.Fatal(err)}
 if err=database.GetDB().Create(&model.User{Username:"xpanel-local-test",Password:hash}).Error;err!=nil{t.Fatal(err)}
 prevMgr:=runtime.GetManager();runtime.SetManager(runtime.NewManager(runtime.LocalDeps{APIPort:func()int{return 0},SetNeedRestart:func(){}}));defer runtime.SetManager(prevMgr)
 bin:=os.Getenv("XPANEL_VISUAL_BIN");if bin==""{t.Fatal("runtime bin required")};t.Setenv("XUI_BIN_FOLDER",bin)
 template,err:=os.ReadFile("service/config.json");if err!=nil{t.Fatal(err)}
 var cfg map[string]any;if err=json.Unmarshal(template,&cfg);err!=nil{t.Fatal(err)}
 freePort:=func()int{l,e:=net.Listen("tcp","127.0.0.1:0");if e!=nil{t.Fatal(e)};defer l.Close();return l.Addr().(*net.TCPAddr).Port}
 cfg["inbounds"].([]any)[0].(map[string]any)["port"]=freePort()
 cfg["metrics"].(map[string]any)["listen"]="127.0.0.1:"+strconv.Itoa(freePort())
 cfg["outbounds"]=[]any{map[string]any{"tag":"blocked","protocol":"blackhole","settings":map[string]any{}}}
 cfg["routing"]=map[string]any{"domainStrategy":"AsIs","rules":[]any{map[string]any{"type":"field","inboundTag":[]string{"api"},"outboundTag":"api"}}}
 origin:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){io.WriteString(w,"XPANEL_ACTUAL_LOOPBACK_PAYLOAD")}));defer origin.Close()
 proxyPort:=freePort()
 cfg["inbounds"]=append(cfg["inbounds"].([]any),map[string]any{"tag":"xpanel-smoke","listen":"127.0.0.1","port":proxyPort,"protocol":"socks","settings":map[string]any{"auth":"password","accounts":[]any{map[string]any{"user":"smoke","pass":"synthetic-smoke"}},"udp":false}})
 cfg["outbounds"]=append(cfg["outbounds"].([]any),map[string]any{"tag":"local-origin","protocol":"freedom","settings":map[string]any{}})
 cfg["routing"].(map[string]any)["rules"]=append(cfg["routing"].(map[string]any)["rules"].([]any),map[string]any{"type":"field","inboundTag":[]string{"xpanel-smoke"},"ip":[]string{"127.0.0.1/32"},"outboundTag":"local-origin"})
 encoded,err:=json.Marshal(cfg);if err!=nil{t.Fatal(err)}
 if err=database.GetDB().Save(&model.Setting{Key:"xrayTemplateConfig",Value:string(encoded)}).Error;err!=nil{t.Fatal(err)}
 xs:=service.XrayService{};if err=xs.RestartXray(true);err!=nil{t.Fatal(err)};defer xs.StopXray()
 var conn net.Conn;for i:=0;i<40;i++{conn,err=net.DialTimeout("tcp",net.JoinHostPort("127.0.0.1",strconv.Itoa(proxyPort)),time.Millisecond*100);if err==nil{break};time.Sleep(time.Millisecond*50)};if err!=nil{t.Fatal(err)};defer conn.Close();conn.SetDeadline(time.Now().Add(time.Second*4))
 conn.Write([]byte{5,1,2});reply:=make([]byte,2);if _,err=io.ReadFull(conn,reply);err!=nil||reply[1]!=2{t.Fatalf("socks greeting %v %v",reply,err)}
 auth:=append([]byte{1,5},[]byte("smoke")...);auth=append(auth,15);auth=append(auth,[]byte("synthetic-smoke")...);conn.Write(auth);if _,err=io.ReadFull(conn,reply);err!=nil||reply[1]!=0{t.Fatal("socks auth",err)}
 _,op,_:=net.SplitHostPort(strings.TrimPrefix(origin.URL,"http://"));port,_:=strconv.Atoi(op);conn.Write([]byte{5,1,0,1,127,0,0,1,byte(port>>8),byte(port)});response:=make([]byte,10);if _,err=io.ReadFull(conn,response);err!=nil||response[1]!=0{t.Fatal("socks connect",err,response)}
 fmt.Fprintf(conn,"GET / HTTP/1.0\r\nHost: local-origin\r\n\r\n");body,err:=io.ReadAll(conn);if err!=nil||!strings.Contains(string(body),"XPANEL_ACTUAL_LOOPBACK_PAYLOAD"){t.Fatalf("actual payload failed %v",err)};t.Log("actual authenticated SOCKS -> pinned core -> loopback HTTP payload passed")
 s:=NewServer(nil);s.cron=cron.New();defer s.cancel();previous:=global.GetWebServer();global.SetWebServer(s);defer global.SetWebServer(previous)
 var restartRequested atomic.Bool;global.SetRestartHook(func(){restartRequested.Store(true)});defer global.SetRestartHook(func(){})
 router,err:=s.initRouter();if err!=nil{t.Fatal(err)}
 server:=httptest.NewServer(router);defer server.Close()
 if err=os.WriteFile(out,[]byte(server.URL),0600);err!=nil{t.Fatal(err)}
 defer os.Remove(out)
 for i:=0;i<900;i++{if _,err:=os.Stat(out+".done");err==nil{if os.Getenv("XPANEL_REQUIRE_RESTORE_RESTART")=="true"&&!restartRequested.Load(){t.Fatal("restore did not request restart")};t.Log("restore restart requested",restartRequested.Load());return};time.Sleep(100*time.Millisecond)}
 t.Fatal("visual client timed out")
}
