package lunch

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"cloud.google.com/go/datastore"
)

type Parameter struct {
	SubCommand string
	Value      string
}

type Restaurant struct {
	ID        int64     'datastore:"-"'
	Name      string    'datastore:"name"'
	CreatedAt time.Time 'datastore:"createdAt"'
}

func Lunch(w http.ResponseWriter, r *http.Request) {
	// POSTのみを許可する
	if r.Method != "POST" {
		e := "Method Not Allowed."
		log.Println(e)
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(e))
		return
	}

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("ReadAllError: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	parsed, err := url.ParseQuery(string(b))
	if err != nil {
		log.Printf("ParseQueryError: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	// token認証処理
	if parsed.Get("token") != os.Getenv("SLACK_TOKEN") {
		e := "Unauthorized Token."
		log.Printf(e)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(e))
		return
	}

	p := new(Parameter)
	p.parse(parsed.Get("text"))

	switch p.SubCommand {
	case "add":
		if err := add(p.Value); err != nil {
			log.Printf("DatastorePutError: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(e))
		}

	case "list":
		// listのしょり

	default:
		e := "Invalid SubCommand."
		log.Println(e)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(e))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(p.Value))
}

func (p *Parameter) parse(text string) {
	// 先頭と末尾の空白を除去
	t := strings.TrimSpace(text)
	if len(t) < 1 {
		return
	}
	// " "で区切った文字列の0番目とそれ以降で構成された,長さ2のスライスを返却する
	s := strings.SplitN(t, " ", 2)
	p.SubCommand = s[0]

	if len(s) == 1 {
		return
	}

	p.Value = s[1]
}

func add(value string) error {
	ctx := context.Background()
	client, err := datastore.NewClient(ctx, os.Getenv("PROJECT_NAME"))
	if err != nil {
		return err
	}

	newKey := datastore.IncompleteKey("Restaurant", nil)
	r := Restaurant {
		Name: value,
		Created: time.Now(),
	}
	if _, err := client.Put(ctx, newKey, $r); err != nil {
		return err
	}
	return nil
}
