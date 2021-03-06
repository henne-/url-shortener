package main

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"regexp"
	"time"

	"github.com/globalsign/mgo/bson"
	"github.com/go-bongo/bongo"
	"github.com/sirupsen/logrus"
)

var charset = []string {"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z", "A", "B", "C", "D", "E", "F", "G", "H", "U", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z"}
var emojiCharset = []string {"😀","😃","😄","😁","😆","😊","😂","😅","😇","🙂","🙃","😉","😌","😚","😙","😗","😘","😍","😋","😛","😝","😜","🤪","🤩","😎","🤓","🧐","🤨","😏","😒","😞","😔","😟","😖","😣","🙁","😕","😫","😩","😢","😭","😤","😠","😡","🤬","🤯","😳","😓","😥","😰","😨","😱","🤗","🤔","🤭","🤫","🤥","🙄","😬","😑","😐","😶","😯","😦","😧","😮","😲","🤐","😵","😪","🤤","😴","🤢","🤮","🤧","😷","🤒","🤕","🤑","🤠","😈","👿","👻","💩","🤡","👺","👹","💀","👽","👾","🤖","😻","😹","😸","😺","🎃","😼","😽","🙀","😿","😾","🙌","🤲","👐","👏","🤝","🤛","✊","👊","👎","👍","🤜","🤞","🤟","🤘","👇","👆","👉","👈","👌","✋","🤚","🖐","🖖","🖕","💪","🤙","👋","🙏","💍","💄","💋","👄","👁","👣","👃","👂","👅","👀","🧠","🗣","👤","👥","👩","👦","🧒","👧","👶","🧑","👨","👱","🧔","👳","👲","👴","🧓","👵","🧕","👮","👷","💂","🕵","🍳","🌾","🎓","🎤","💻","🏭","🏫","💼","🔧","🚒","🎨","🔬","🚀","👸","🤵","👰","🤴","🤶","🎅","🧙","🧟","🧛","🧝","🧜","👗","👔","👖","👕","👚","👙","👘","👠","👡","👢","🧣","🧤","🧦","👟","👞","🎩","🧢","👒","⛑","👜","👛","👝","👑","🎒","👓","🕶","🌂","🐶","🐱","🐭","🐹","🐰","🐯","🐨","🐼","🐻","🦊","🦁","🐮","🐷","🐽","🐸","🐒","🙊","🙉","🙈","🐵","🐔","🐧","🐦","🐤","🐣","🦇","🦉","🦅","🦆","🐥","🐺","🐗","🐴","🦄","🐝","🐞","🐚","🐌","🦋","🐛","🐜","🦗","🕷","🕸","🦂","🦕","🦖","🦎","🐍","🐢","🐙","🦑","🦐","🦀","🐡","🐋","🐳","🐬","🐟","🐠","🦈","🐊","🐅","🐆","🦓","🐫","🐪","🦏","🐘","🦍","🦒","🐃","🐂","🐄","🐎","🦌","🐐","🐑","🐏","🐖","🐕","🐩","🐈","🐓","🦃","🐿","🐀","🐁","🐇","🕊","🦔","🐾","🐉","🐲","🌵","🌱","🌴","🌳","🌲","🎄","🌿","☘","🍀","🎍","🎋","🍄","🍁","🍂","🍃","💐","🌷","🌹","🥀","🌺","🌝","🌞","🌻","🌼","🌸","🌛","🌜","🌚","🌕","🌖","🌓","🌒","🌑","🌘","🌗","🌔","🌙","🌎","🌍","🌏","✨","🌟","⭐","💫","💥","🔥","🌪","🌈","🌥","⛅","🌤","🌦","🌧","⛈","🌩","🌨","💨","🌬","⛄","💧","💦","🌊","🌫","🍏","🍎","🍐","🍊","🍋","🍈","🍓","🍇","🍉","🍌","🍒","🍑","🍍","🥥","🥝","🥒","🥦","🥑","🍆","🍅","🌶","🌽","🥕","🥔","🍠","🧀","🥨","🥖","🍞","🥐","🥚","🥞","🥓","🥩","🍟","🍔","🌭","🍖","🍗","🍕","🥪","🥙","🌮","🌯","🥗","🥘","🥫","🍝","🍜","🥟","🍱","🍣","🍛","🍲","🍤","🍙","🍚","🍘","🍥","🍨","🍧","🍡","🍢","🥠","🍦","🥧","🍰","🎂","🍮","🍩","🍿","🍫","🍬","🍭","🍪","🌰","🥜","🍯","🥛","🍶","🥤","🍼","🍺","🍻","🥂","🍷","🥃","🍴","🥄","🍾","🍹","🍸","🍽","🥣","🥡","🥢","🎾","🏈","🏀","⚽","🏐","🏉","🎱","🏓","🏸","⛳","🏏","🏑","🏒","🥅","🏹","🎣","🥊","🎽","⛷","🎿","🛷","🥌","⛸","🏂","🤼","🏆","🥇","🥈","🥉","🎫","🎗","🏵","🎖","🏅","🎟","🎪","🎭","🎼","🎧","🎬","🎹","🥁","🎷","🎺","🎸","🎮","🎳","🎯","🎲","🎻","🎰","🚗","🚕","🚙","🚌","🚎","🚐","🚑","🚓","🏎","🚚","🚛","🚜","🛴","🚲","🚍","🚔","🚨","🏍","🛵","🚘","🚖","🚡","🚠","🚄","🚝","🚞","🚋","🚃","🚅","🚈","🚂","🚇","🛬","🛫","🚉","🚊","🛩","💺","🛰","🛸","🛳","⛴","🚢","⛽","🗺","🚏","🚥","🚦","🚧","🗿","🗽","🗼","🏰","🏯","⛲","🎠","🎢","🎡","🏟","⛱","🏖","🏝","🏜","⛺","🏕","🗻","🏔","⛰","⌚","📱","🖥","🖨","🗜","💽","📼","📷","📽","🎞","📠","📺","🎛","⏱","🔦","💡","📞","📸","💶","💴","💵","💸","🔫","💣","🚬","🎊","🎉","🏮","🎏","💊","🚿","🗝","🚰","🔑","🛎","🛏","🛒", "❤️", "🧡","💛","💚","💙","💜","🖤","💯","✅","❎","💤","❔","❓","❕","❗"}

func getUniqueRandomString(length int, emoji bool) (string, error) {
Start:
	name := randomString(length, emoji)
	link := &Link{}
	err := connection.Collection("links").FindOne(bson.M{"Name": name}, link)
	if _, ok := err.(*bongo.DocumentNotFoundError); ok {
		return name, nil
	} else if err != nil {
		return "", err
	} else {
		goto Start
	}
}

func randomString(length int, emoji bool) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomString := ""
	if emoji {
		for i := 0; i < length; i++ {
			randomString += emojiCharset[r.Intn(len(emojiCharset))]
		}
	} else {
		for i := 0; i < length; i++ {
			randomString += charset[r.Intn(len(charset))]
		}
	}
	return randomString
}


func getLink(name string) (error, *Link) {
	link := &Link{}
	err := connection.Collection("links").FindOne(bson.M{"name": name}, link)
	if _, ok := err.(*bongo.DocumentNotFoundError); ok {
		logrus.Info(fmt.Sprintf("Short \"%s\" not found", name))
		return errors.New("404"), nil
	}
	return err, link
}
func addHttp(url string) string {
	r := regexp.MustCompile("^(((f|ht)tps?)|tg|steam)://")
	if !r.MatchString(url) {
		url = "http://" + url
	}
	return url
}

func returnError500(err error, w http.ResponseWriter) {
	logrus.Error(err)
	w.WriteHeader(http.StatusInternalServerError)
	_, _ = fmt.Fprintf(w, "Internal Server error..")
}

func returnError404(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
	_, _ = fmt.Fprintf(w, "404 Document not found.")
}
func returnError401(w http.ResponseWriter) {
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = fmt.Fprintf(w, "Unauthorized.")
}
