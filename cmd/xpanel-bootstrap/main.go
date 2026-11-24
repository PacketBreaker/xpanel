// xpanel-bootstrap initializes only a new, private portable-install database.
package main

import (
 "fmt"
 "os"
 "path/filepath"
 "github.com/mhsanaei/3x-ui/v3/internal/config"
 "github.com/mhsanaei/3x-ui/v3/internal/database"
 "github.com/mhsanaei/3x-ui/v3/internal/database/model"
)

func main(){
 db:=config.GetDBPath()
 if _,err:=os.Stat(db);err==nil{owned,e:=os.ReadFile(db+".xpanel-owned");if e!=nil||string(owned)!="Xpanel portable v1\n"{fail(fmt.Errorf("existing database is not owned by this portable installation; refusing adoption or migration"))};fmt.Fprintln(os.Stderr,"Owned database left unchanged. Use Panel Settings to change listeners.");return}else if !os.IsNotExist(err){fail(err)}
 if err:=os.MkdirAll(filepath.Dir(db),0700);err!=nil{fail(err)}
 // O_EXCL is an intent marker, not an attempt to adopt an existing installation.
 marker:=db+".bootstrap-lock";f,err:=os.OpenFile(marker,os.O_CREATE|os.O_EXCL|os.O_WRONLY,0600);if err!=nil{fail(err)};f.Close();defer os.Remove(marker)
 if _,err=os.Stat(db);err==nil{fail(fmt.Errorf("database appeared during initialization"))}
 if err=database.InitDB(db);err!=nil{fail(err)};defer database.CloseDB()
 rows:=[]model.Setting{{Key:"webListen",Value:"127.0.0.1"},{Key:"webPort",Value:"2053"},{Key:"subListen",Value:"127.0.0.1"},{Key:"subEnable",Value:"false"},{Key:"subJsonEnable",Value:"false"},{Key:"subTitle",Value:"Xpanel"}}
 if err=database.GetDB().Save(&rows).Error;err!=nil{fail(err)}
 if err=os.WriteFile(db+".xpanel-owned",[]byte("Xpanel portable v1\n"),0600);err!=nil{fail(err)}
 fmt.Println("New private Xpanel database initialized. Panel binds 127.0.0.1:2053; subscriptions disabled until configured. No existing data imported.")
}
func fail(err error){fmt.Fprintln(os.Stderr,err);os.Exit(1)}
