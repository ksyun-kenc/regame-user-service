package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"strings"

	"regame-user-service/database"
	"regame-user-service/service"
)

func main() {
	pdAddr := os.Getenv("PD_ADDR")
	if pdAddr != "" {
		os.Args = append(os.Args, "-pd", pdAddr)
	}
	pdAddress := flag.String("pd", "tikv://127.0.0.1:2379", "pd address")
	option := flag.String("op", "", "Option: add, del, get, list, set")
	flag.Parse()

	if pdAddress == nil || len(*pdAddress) == 0 || option == nil || len(*option) == 0 {
		flag.CommandLine.Usage()
		return
	}

	switch *option {
	case "add":
		if flag.NArg() < 2 {
			fmt.Println("Usage: add <username> <password>")
			return
		}
		passwordStr := flag.Args()[1]
		if len(passwordStr) != service.SM3PasswordLength {
			fmt.Println("Invalid password:", passwordStr)
			return
		}
		password, err := hex.DecodeString(passwordStr)
		if err != nil {
			fmt.Println("Invalid password:", passwordStr)
			return
		}
		database.TiKVClientInit(*pdAddress)
		usernameKey := database.PASSWORD_KEY + flag.Args()[0]
		err = database.TiKVClientPuts([]byte(usernameKey), password)
		if err != nil {
			fmt.Println("Puts failed:", err)
			return
		}
	case "del", "delete", "rm", "remove":
		if flag.NArg() < 1 {
			fmt.Println("Usage: del <username>...")
			return
		}
		database.TiKVClientInit(*pdAddress)
		usernames := make([][]byte, 0)
		for _, username := range flag.Args() {
			usernames = append(usernames, []byte(database.PASSWORD_KEY+username))
		}
		err := database.TiKVClientDeletes(usernames...)
		if err != nil {
			fmt.Println("Deletes failed:", err)
			return
		}
	case "get":
		if flag.NArg() < 1 {
			fmt.Println("Usage: get <username>")
			return
		}
		database.TiKVClientInit(*pdAddress)

		username := flag.Args()[0]
		kv, err := database.TiKVClientGet([]byte(database.PASSWORD_KEY + username))
		if err != nil {
			fmt.Println("Get failed:", err)
			return
		}
		fmt.Printf("%s: %s\n", username, hex.EncodeToString(kv.V))
	case "list":
		database.TiKVClientInit(*pdAddress)
		kvs, err := database.TiKVClientScan([]byte(database.PASSWORD_KEY), 1024)
		if err != nil {
			fmt.Println("Scan failed:", err)
			return
		}
		for i, kv := range kvs {
			fmt.Printf("[%d] %s: %s\n", i, strings.TrimLeft(string(kv.K), database.PASSWORD_KEY), hex.EncodeToString(kv.V))
		}
	case "set":
		if flag.NArg() < 2 {
			fmt.Println("Usage: set <username> <password>")
			return
		}
		passwordStr := flag.Args()[1]
		if len(passwordStr) != service.SM3PasswordLength {
			fmt.Println("Invalid password:", passwordStr)
			return
		}
		password, err := hex.DecodeString(passwordStr)
		if err != nil {
			fmt.Println("Invalid password:", passwordStr)
			return
		}
		database.TiKVClientInit(*pdAddress)
		username := flag.Args()[0]
		oldValue, err := database.TiKVClientUpdate([]byte(database.PASSWORD_KEY+username), password)
		if err != nil {
			fmt.Println("Update failed:", err)
			return
		}
		fmt.Printf("%s: %s -> %s\n", username, hex.EncodeToString(oldValue), passwordStr)
	case "addcode":
		if flag.NArg() < 1 {
			fmt.Println("Usage: addcode <username>")
			return
		}
		// TODO: Generate a One-time password [000000,999999] and deadline time.Now()+10min
		// Key=CODE_KEY+username, Value=password_deadline
	default:
		flag.CommandLine.Usage()
		return
	}
}
