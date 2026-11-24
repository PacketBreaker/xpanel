//go:build linux

package panel

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"time"
)

func channelRoot() (string, error) {
	root := os.Getenv("XPANEL_MANAGED_ROOT")
	if root == "" || !filepath.IsAbs(root) {
		return "", fmt.Errorf("Xpanel updates require the managed supervisor and an operator-configured trusted channel")
	}
	data, err := os.ReadFile(filepath.Join(root, "supervisor.json"))
	if err != nil {
		return "", fmt.Errorf("Xpanel supervisor unavailable")
	}
	var h struct {
		At int64 `json:"at"`
	}
	if json.Unmarshal(data, &h) != nil || time.Now().Unix()-h.At > 10 || h.At > time.Now().Unix()+2 {
		return "", fmt.Errorf("Xpanel supervisor heartbeat is stale")
	}
	return root, nil
}
func channelPreview(root string) (map[string]any, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 55*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "python3", filepath.Join(root, "host-tools", "xpanel-supervisor.py"), "preview", "--root", root)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("Xpanel trusted channel is unavailable or verification failed")
	}
	var j map[string]any
	if json.Unmarshal(out, &j) != nil {
		return nil, fmt.Errorf("invalid channel preview")
	}
	return j, nil
}
func brandedUpdateInfo() (*PanelUpdateInfo, error) {
	root, err := channelRoot()
	if err != nil {
		return nil, err
	}
	j, err := channelPreview(root)
	if err != nil {
		return nil, err
	}
	current := "Xpanel"
	if b, e := os.ReadFile(filepath.Join(root, "CHANNEL.json")); e == nil {
		var receipt map[string]any
		if json.Unmarshal(b, &receipt) == nil {
			current = fmt.Sprintf("Xpanel sequence %.0f", receipt["sequence"])
		}
	}
	return &PanelUpdateInfo{Channel: "xpanel", CurrentVersion: current, LatestVersion: fmt.Sprintf("Xpanel sequence %.0f", j["sequence"]), UpdateAvailable: j["update_available"] == true, ManifestSHA256: fmt.Sprint(j["manifest_sha256"])}, nil
}
func startBrandedUpdate(useDev bool) (int64, error) {
	return 0, fmt.Errorf("Check the signed channel in this browser before confirming an update")
}
func (s *PanelService) StartXpanelUpdate(expected string) (int64, error) {
	root, err := channelRoot()
	if err != nil {
		return 0, err
	}
	lock, err := os.OpenFile(filepath.Join(root, ".update-request.lock"), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return 0, err
	}
	defer lock.Close()
	if err = syscall.Flock(int(lock.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		return 0, fmt.Errorf("an update is active")
	}
	j, err := channelPreview(root)
	if err != nil {
		return 0, err
	}
	if expected == "" || expected != j["manifest_sha256"] || j["update_available"] != true {
		return 0, fmt.Errorf("Channel changed or no update available; check again")
	}
	if _, err := os.Stat(filepath.Join(root, "update-active.json")); err == nil {
		return 0, fmt.Errorf("an Xpanel update is active")
	}
	run := time.Now().UnixNano()
	job := map[string]any{"runId": strconv.FormatInt(run, 10), "manifest_sha256": j["manifest_sha256"]}
	data, _ := json.Marshal(job)
	path := filepath.Join(root, "update-request.json")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return 0, fmt.Errorf("an Xpanel update is already queued")
	}
	if _, err = f.Write(data); err != nil {
		f.Close()
		os.Remove(path)
		return 0, err
	}
	err = f.Sync()
	f.Close()
	if err != nil {
		os.Remove(path)
		return 0, err
	}
	return run, nil
}
