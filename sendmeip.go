package main

import (
	"fmt"
	"log"
	"net/smtp"
	"regexp"
	"strings"
	"time"

	"github.com/chyeh/pubip"
	"github.com/patrickmn/go-cache"
	"github.com/spf13/viper"
)

func init() {
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/sendmeip/")
	viper.AddConfigPath("$HOME/.sendmeip/")
	viper.AddConfigPath("./")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	viper.SetDefault("refreshInterval", 35)
	viper.SetDefault("mailSubject", "New IP")
	viper.SetDefault("smtpHost", "smtp.gmail.com")
	viper.SetDefault("smtpPort", 587)
	viper.SetDefault("smtpUser", "user@example.org")
	viper.SetDefault("smtpPass", "abc123")
	viper.SetDefault("notifyAddr", "notify@example.org")
}

func main() {
	t := time.NewTicker(time.Duration(viper.GetInt64("refreshInterval")) * time.Minute)
	c := cache.New(5*time.Hour, 30*time.Second)
	fmt.Printf("Running ...\n(press CTRL + C to quit)")
	for now := range t.C {
		x, found := c.Get("ipa")
		if !found {
			x = "NOTFOUND"
		}
		ip, _ := pubip.Get()
		ipAddr := fmt.Sprintf("%s", ip)
		if ipAddr != x.(string) && validIP4(ipAddr) {
			c.Set("ipa", ipAddr, cache.NoExpiration)
			send(fmt.Sprintf("New IP Address at home.\nIP: %s\nChanged: %s\n", ip, now))
		}
	}
}

func send(body string) {
	from := viper.GetString("smtpUser")
	pass := viper.GetString("smtpPass")
	to := viper.GetString("notifyAddr")

	msg := fmt.Sprintf("From: %s\nTo: %s\nSubject: %s\n\n%s", from, to, viper.GetString("mailSubject"), body)

	err := smtp.SendMail(fmt.Sprintf("%s:%d", viper.GetString("smtpHost"), viper.GetInt("smtpPort")),
		smtp.PlainAuth("", from, pass, viper.GetString("smtpHost")),
		from, []string{to}, []byte(msg))

	if err != nil {
		log.Printf("smtp error: %s", err)
		return
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Println(err.Error())
	}
}

func validIP4(ipAddress string) bool {
	ipAddress = strings.Trim(ipAddress, " ")

	re, _ := regexp.Compile(`^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$`)
	if re.MatchString(ipAddress) {
		return true
	}
	return false
}
