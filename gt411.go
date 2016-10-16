package main

import (
  "fmt"
  "net/http"
  "net/url"
  "io/ioutil"
  "encoding/json"
  "os"
  "github.com/urfave/cli"
  "strconv"
  "text/tabwriter"
  "strings"

)


type Credentials struct {Token string}
type Configuration struct {
  Username   string
  Password   string
}

type Torrent struct{
  Id string
  Name string
  Size string
  Category string
}

//constants

const baseUrl = "https://api.t411.ch"
const padding = 2

func main() {
  app := cli.NewApp()
  app.Name = "GT411"
  app.Usage = "CLI for T411 API"
  app.Version = "0.0.1"
  app.Commands = []cli.Command{
    {
      Name:    "top",
      Aliases: []string{"t"},
      Usage:   "Display top downloads per day/week/month/100",
      Flags: []cli.Flag{
          cli.StringFlag{
              Name:  "c",
              Value: "null",
              Usage: "Name of the category.",
          },
      },
      Action:  func(c *cli.Context) error {
        switch c.Args().First(){
        case "today","month","week","100":
          t := torrents("/torrents/top/"+c.Args().First(),auth().Token)
          if c.String("c") != "null"{
            printTable(sortedOnCategory(category()[c.String("c")],t))
          }else{
            printTable(t)
           }
        default:
          fmt.Println("don't understand: ", c.Args().First())
        }
        return nil
      },
    },
    {
    Name: "search",
    Aliases: []string{"s"},
    Flags: []cli.Flag{
        cli.StringFlag{
            Name:  "c",
            Value: "null",
            Usage: "Name of the category.",
        },
        cli.StringFlag{
            Name:  "l",
            Value: "null",
            Usage: "Limit of the amount of records",
        },
    },
    Usage:  "Search for torrents",
    Action:  func(c *cli.Context) error {
      arr := make([]string, 0)
      var query string
      if c.String("c") != "null"{ arr = append(arr,"cid="+category()[c.String("c")]) }
      if c.String("l") != "null"{ arr = append(arr,"limit="+c.String("l")) }
      for _,v := range arr { query = query + "&" +v }
      t := torrents("/torrents/search/"+c.Args().First()+query,auth().Token)
      fmt.Println(t)
      //printTable(t)
      return nil
      },
    },
  }
  app.Run(os.Args)
}

func config() (c *Configuration) {
  file, _ := os.Open(os.Getenv("HOME") + "/.config/go-t411.json")
  decoder := json.NewDecoder(file)
  err := decoder.Decode(&c)
  if err != nil {
    fmt.Println("error:", err)
  }
  return
}

func auth() (c *Credentials) {
  configuration := config()
  resp, err := http.PostForm(baseUrl + "/auth", url.Values{"username": {configuration.Username}, "password": {configuration.Password}})
  if err != nil {
	   fmt.Println(err)
     os.Exit(1)
  }
  defer resp.Body.Close()
  body, err := ioutil.ReadAll(resp.Body)
  jerr := json.Unmarshal(body, &c)
  if jerr != nil {
	   fmt.Println(string(body))
	}
	return
}

func torrents(action,token string) (data []Torrent){
  //configuration := config()
  uri := baseUrl + action
  torrents := make([]Torrent, 0)
  client := &http.Client{}
  req, err := http.NewRequest("GET",uri,nil)
  req.Header.Add("Authorization",token)
  resp, err := client.Do(req); if err != nil {
    fmt.Println("HTTP connection error");fmt.Println(err)
    os.Exit(1)
  }
  defer resp.Body.Close()
  body, err := ioutil.ReadAll(resp.Body); if err != nil {
    fmt.Println("Body read error");fmt.Println(err)
    os.Exit(1)
  }

  if strings.Contains(action,"/torrents/search"){
      //TODO
      fmt.Println(string(body))
      //os.Exit(1)

  }else{
    jerr := json.Unmarshal(body,&torrents); if jerr != nil {
      fmt.Println(jerr)
      os.Exit(1)
      }
  }

  for _, v := range torrents {
    i, _ := strconv.Atoi(v.Size);i = i / 1000000
    s := strconv.Itoa(i); v.Size = s + " Mb"
    data = append(data,v)
  }
  return
}

func category()(c map[string]string){
  c = make(map[string]string)
  c["movie"] = "631"
  c["serie"] = "433"
  return
}

func printTable(data []Torrent){
  w := tabwriter.NewWriter(os.Stdout, 0, 0, padding, ' ', tabwriter.Debug)
  for _, v := range data{
    fmt.Fprintln(w, v.Id + "\t" + v.Name + "\t" + v.Size + "\t")
  }
	w.Flush()
}

func sortedOnCategory(c string, u []Torrent)(f []Torrent){
  for _, v := range u {
    if c == v.Category { f = append(f,v) }
  }
  return
}
