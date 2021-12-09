package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/Jeffail/gabs/v2"
)

func checkErr(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "An error occurred: %s\n", err.Error())
		os.Exit(1)
	}
}

func findServerID(url string) string {
	res, err := http.Get(url)
	checkErr(err)

	j, err := gabs.ParseJSONBuffer(res.Body)
	checkErr(err)
	res.Body.Close()

	hostname := j.Children()[0].Search("hostname").Data().(string)
	return strings.Split(hostname, ".")[0]
}

func main() {
	country := flag.String("country", "", "Country id")
	serverNumber := flag.Int("server", -1, "Server number")
	protocol := flag.String("protocol", "tcp", "Protocol (tcp/udp)")
	configFolder := flag.String("config", "", "Config folder")
	credentialFile := flag.String("creds", "", "Credential file")
	flag.Parse()

	if *configFolder == "" || *credentialFile == "" {
		checkErr(fmt.Errorf("Invalid parameter"))
	}

	if !strings.EqualFold(*protocol, "tcp") && !strings.EqualFold(*protocol, "udp") {
		checkErr(fmt.Errorf("Invalid protocol: %s", *protocol))
	}

	serverID := ""

	if *country == "" {

		if *serverNumber != -1 {
			checkErr(fmt.Errorf("Country is not set but server number is"))
		}

		serverID = findServerID("https://nordvpn.com/wp-admin/admin-ajax.php?action=servers_recommendations")
	} else {
		if *serverNumber != -1 {
			serverID = *country + strconv.FormatInt(int64(*serverNumber), 10)
		} else {
			res, err := http.Get("https://nordvpn.com/wp-admin/admin-ajax.php?action=servers_countries")
			checkErr(err)

			j, err := gabs.ParseJSONBuffer(res.Body)
			checkErr(err)
			res.Body.Close()

			countryID := -1
			for _, c := range j.Children() {
				cc := c.Search("code").Data().(string)
				if strings.EqualFold(cc, *country) {
					countryID = int(c.Search("id").Data().(float64))
				}
			}

			if countryID == -1 {
				checkErr(fmt.Errorf("Invalid country: %s", *country))
			}

			serverID = findServerID(
				fmt.Sprintf(
					"https://nordvpn.com/wp-admin/admin-ajax.php?action=servers_recommendations&filters={%%22country_id%%22:%d}",
					countryID,
				),
			)
		}
	}

	fmt.Printf("sudo openvpn --config \"%s/%s.nordvpn.com.%s.ovpn\" --auth-user-pass \"%s\"\n",
		*configFolder,
		serverID, *protocol,
		*credentialFile,
	)

}
