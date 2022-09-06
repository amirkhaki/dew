package cmd

import (
	"cf/mashup"
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/briandowns/spinner"
	"github.com/jaxleof/uispinner"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var div3Diffculty = [][]string{{"800"}, {"800", "900"}, {"900", "1000", "1100"}, {"1100", "1200", "1300", "1400"}, {"1400", "1500", "1600", "1700"}, {"1700", "1800", "1900"}, {"1900", "2000", "2100"}}
var div2Diffculty = [][]string{{"800", "900", "1000"}, {"1000", "1100", "1200"}, {"1200", "1300", "1400", "1500", "1600"}, {"1600", "1700", "1800", "1900"}, {"2000", "2100", "2200", "2300", "2400"}, {"2500", "2600", "2700", "2800"}}
var div1Diffculty = [][]string{{"1500", "1600", "1700"}, {"1800", "1900", "2000", "2100", "2200", "2300"}, {"2400", "2500", "2700", "2800"}, {"2900", "3000", "3100", "3200", "3300", "3400", "3500"}, {"3400", "3500"}}

func init() {
	rootCmd.AddCommand(NewCmd)
	NewCmd.AddCommand(div1)
	NewCmd.AddCommand(div2)
	NewCmd.AddCommand(div3)
	NewCmd.AddCommand(randomOne)
}

const (
	title    = "miaonei"
	duration = "120"
)

var NewCmd = &cobra.Command{
	Use:   "generate",
	Short: "create a contest",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}
var div3 = &cobra.Command{
	Use:   "div3",
	Short: "create a contest, whose difficulty like div3",
	Run: func(cmd *cobra.Command, args []string) {
		newContest(div3Diffculty)
	},
}
var div2 = &cobra.Command{
	Use:   "div2",
	Short: "create a contest, whose difficulty like div2",
	Run: func(cmd *cobra.Command, args []string) {
		newContest(div2Diffculty)
	},
}
var div1 = &cobra.Command{
	Use:   "div1",
	Short: "create a contest, whose difficulty like div1",
	Run: func(cmd *cobra.Command, args []string) {
		newContest(div1Diffculty)
	},
}

func newContest(diffculty [][]string) {
	pro := PickSomeProblem(diffculty)
	mashup.Login()
	mashup.CreateContest(title, duration, pro)
	OpenWebsite("https://codeforces.com/mashups")
}

var randomOne = &cobra.Command{
	Use:   "random",
	Short: "random select one problem",
	Run: func(cmd *cobra.Command, args []string) {
		Random()
	},
}

func Random() {
	isExist := checkConfigFile()
	if !isExist {
		log.Fatal("config file is not exist, please use cf init command")
	}
	viper.SetConfigFile("./codeforces/config.yaml")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal(err)
	}
	if !viper.IsSet("rating") {
		log.Fatal("we notice the info of rating is not exist, please use cf init config command first, or modify rating in ./codeforces/config.yaml (you can add a line and write 'rating: 1234').")
	}
	var rating = viper.GetInt("rating")
	if rating < 800 {
		rating = 800
	}
	rating = (rating / 100) * 100
	lowRating := rating + 200
	if lowRating > 3500 {
		lowRating = 3500
	}
	highRating := rating + 500
	if highRating > 3500 {
		highRating = 3500
	}
	var pro []string
	for i := lowRating; i <= highRating; i += 100 {
		pro = append(pro, strconv.Itoa(i))
	}
	var thisOne = PickOneProblem(pro)
	viper.Set("random", strconv.Itoa(thisOne.ContestId)+thisOne.Index)
	err = viper.WriteConfig()
	if err != nil {
		log.Fatal(err)
	}
	OpenRandomFunc(thisOne)
}

func PickSomeProblem(in [][]string) []string {
	cj := uispinner.New()
	cj.Start()
	login := cj.AddSpinner(spinner.CharSets[34], 100*time.Millisecond).SetPrefix("picking problems").SetComplete("pick problem complete")
	var pro []string
	var mp map[string]bool = make(map[string]bool)
	for i := 0; i < len(in); i++ {
		var one = PickOneProblem(in[i])
		var goal = strconv.Itoa(one.ContestId) + one.Index
		if mp[goal] {
			i--
			continue
		}
		pro = append(pro, goal)
		mp[goal] = true
	}
	login.Done()
	cj.Stop()
	return pro
}

func PickOneProblem(r []string) problemInfo {
	data := PickProblems(r)
	data = Deduplication(data, mashup.GetStatus())
	if len(data) == 0 {
		log.Fatal("you are so good, you have solve all problems of the range", r)
	}
	rand.Seed(time.Now().Unix())
	var pos = rand.Int() % len(data)
	return data[pos]
}

func PickProblems(in []string) []problemInfo {
	var res []problemInfo
	for i := 0; i < len(in); i++ {
		data, err := ioutil.ReadFile("./codeforces/" + in[i] + ".json")
		if err != nil {
			log.Fatal(err.Error() + "\nyou should use cf update before generate")
		}
		var tmp []problemInfo
		json.Unmarshal(data, &tmp)
		res = append(res, tmp...)
	}
	return res
}

func Deduplication(data []problemInfo, s map[string]bool) []problemInfo {
	var res []problemInfo
	for i := 0; i < len(data); i++ {
		if _, exist := s[strconv.Itoa(data[i].ContestId)+data[i].Index]; !exist {
			res = append(res, data[i])
		}
	}
	return res
}
