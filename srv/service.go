package service

import (
	"bufio"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"mtgengine/engine"
	srvpb "mtgengine/proto"
	"os"
	"sort"
	"strings"
	"strconv"
)

type Service struct {
	// TODO change to actual db at some point
	cards map[string]*engine.Card
	cols map[string][]*engine.Card
}

func NewService(cards map[string]*engine.Card) *Service {
	return &Service{
		cards: cards,
		cols: make(map[string][]*engine.Card),
	}
}

func (s *Service) CreateCollection(ctx context.Context, req *srvpb.CreateCollectionRequest) (*srvpb.CreateCollectionReply, error) {
	respb := &srvpb.CreateCollectionReply{}
	if _, ok := s.cols[req.Name]; ok {
		return respb, fmt.Errorf("collection %v already exists", req.Name)
	}
	s.cols[req.Name] = make([]*engine.Card, 0)
	return respb, nil
}

func (s *Service) AddCards(ctx context.Context, req *srvpb.AddCardsRequest) (*srvpb.AddCardsReply, error) {
	respb := &srvpb.AddCardsReply{}
	col := req.Collection
	var err error
	var ex *engine.Card
	c := make([]*engine.Card, 0)
	for _, card := range req.Cards {
		if req.ExactMatch {
			ex, err = s.searchExact(card)
			c = append(c, ex)
		} else {
			c, err = s.searchPrefix(card)
		}
		if err != nil || len(c) == 0 {
			respb.UnknownCards = append(respb.UnknownCards, card)
		} else {
			log.Printf("adding card %v to collection %v", c[0].Name, col)
			s.cols[col] = append(s.cols[col], c[0])
		}
	}
	return respb, nil
}

func (s *Service) RemoveCards(ctx context.Context, req *srvpb.RemoveCardsRequest) (*srvpb.RemoveCardsReply, error) {
	respb := &srvpb.RemoveCardsReply{}
	col := req.Collection
	c, ok := s.cols[col]
	if !ok {
		return nil, fmt.Errorf("collection %v doesn't exist\n", col)
	}
	unknown := make([]string, 0)
	for _, card := range req.Cards {
		rem := -1
		for i, ccard := range c {
			if ccard.Name == card {
				rem = i
				break
			}
		}
		if len(c) >= 0 && rem >= 0 {
			c = append(c[:rem], c[rem+1:]...)
			s.cols[col] = c
		} else {
			unknown = append(unknown, card)
		}
	}
	return respb, nil
}

func (s *Service) SearchCards(ctx context.Context, req *srvpb.SearchCardsRequest) (*srvpb.SearchCardsReply, error) {
	respb := &srvpb.SearchCardsReply{}
	switch req.FieldToSearch {
	case srvpb.SearchField_NAME:
		cards, err := s.searchPrefix(req.SearchString)
		for _, c := range cards {
			respb.Cards = append(respb.Cards, struct2proto(c))
		}
		if err != nil {
			return nil, err
		}
	case srvpb.SearchField_TEXT:
		cards, err := s.searchText(req.SearchString)
		if err != nil {
			return nil, err
		}
		for _, c := range cards {
			respb.Cards = append(respb.Cards, struct2proto(c))
		}
	default:
		return nil, fmt.Errorf("unknown search type: %v", req.FieldToSearch)
	}
	return respb, nil
}

func (s *Service) SearchCollection(ctx context.Context, req *srvpb.SearchCollectionRequest) (*srvpb.SearchCollectionReply, error) {
	respb := &srvpb.SearchCollectionReply{}
	cards, err := s.searchName(req.Collection, req.SearchString)
	if err != nil {
		return nil, err
	}
	for _, c := range cards {
		respb.Cards = append(respb.Cards, struct2proto(c))
	}
	return respb, nil
}

func struct2proto(card *engine.Card) *srvpb.Card {
	return &srvpb.Card{
		Artist: card.Artist,
		Cmc: int64(card.Cmc),
		Colors: card.Colors,
		Flavor: card.Flavor,
		Id: card.Id,
		Name: card.Name,
		Power: card.Power,
		Toughness: card.Toughness,
		Rarity: card.Rarity,
		ManaCost: card.ManaCost,
		Subtypes: card.Subtypes,
		Supertypes: card.Supertypes,
		Text: card.Text,
		CType: card.CType,
		Types: card.Types,
	}
}

func (s *Service) searchPrefix(pre string) ([]*engine.Card, error) {
	pre = strings.ToUpper(pre)
	cards := []*engine.Card{}
	for name, card := range s.cards {
		if strings.HasPrefix(name, pre) {
			cards = append(cards, card)
		}
	}
	if len(cards) > 0 {
		return cards, nil
	}
	return nil, fmt.Errorf("no card with prefix: %v", pre)
}

func (s *Service) searchName(col, txt string) ([]*engine.Card, error) {
	txt = strings.ToUpper(txt)
	cards := []*engine.Card{}
	colc, ok := s.cols[col]
	if !ok {
		return nil, fmt.Errorf("no collection: %v", col)
	}
	for _, card := range colc {
		if strings.Contains(strings.ToUpper(card.Name), txt) {
			cards = append(cards, card)
		}
	}
	if len(cards) > 0 {
		return cards, nil
	}
	return nil, fmt.Errorf("no card with %v in the name", txt)
}

func (s *Service) searchExact(str string) (*engine.Card, error) {
	str = strings.ToUpper(str)
	for name, card := range s.cards {
		if str == name {
			return card, nil
		}
	}
	return nil, fmt.Errorf("no card named: %v", str)
}

func (s *Service) searchText(t string) ([]*engine.Card, error) {
	t = strings.ToUpper(t)
	cards := []*engine.Card{}
	for _, card := range s.cards {
		if strings.Contains(strings.ToUpper(card.Text), t) {
			cards = append(cards, card)
		}
	}
	return cards, nil
}

func (s *Service) ExportCollection(ctx context.Context, req *srvpb.ExportCollectionRequest) (*srvpb.ExportCollectionReply, error) {
	respb := &srvpb.ExportCollectionReply{}
	f, err := os.Create(req.FileName)
	if err != nil {
		return respb, err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	col, ok := s.cols[req.Name]
	if !ok {
		return respb, fmt.Errorf("no collection named: %v", req.Name)
	}
	w.Write([]string{"Name", "cmc", "colors", "power", "toughness", "mana cost", "text", "type", "subtypes"})
	for _, c := range col {
		r := csvRow(c)
		w.Write(r)
	}
	w.Flush()
	return respb, nil
}

func csvRow(c *engine.Card) []string {
	//{"Name", "cmc", "colors", "power", "toughness", "mana cost", "text", "type", "subtypes"})
	var record []string
	record = append(record, c.Name)
	record = append(record, strconv.Itoa(c.Cmc))
	record = append(record, color(c.Colors))
	record = append(record, c.Power)
	record = append(record, c.Toughness)
	record = append(record, c.ManaCost)
	record = append(record, c.Text)
	record = append(record, fmt.Sprintf("%v", c.Types))
	record = append(record, fmt.Sprintf("%v", c.Subtypes))
	return record
}

func color(c []string) string {
	if len(c) == 0 {
		return "C"
	}
	// TODO eww
	ks := []string{"W", "U", "B", "R", "G"}
	m := map[string]string{
		"W": "",
		"U": "",
		"B": "",
		"R": "",
		"G": "",
	}
	for _, cl := range c {
		clU := strings.ToUpper(cl)
		m[clU] = clU
	}
	out := ""
	for _, k := range ks {
		out += m[k]
	}
	return out
}

type ScryfallSearchJSON struct {
	Object string `json:"object"`
	Data []map[string]interface{} `json:"data"`
	Status int `json:"status"`
	Details string `json:"details"`
}


func (s *Service) GetPrice(ctx context.Context, req *srvpb.GetPriceRequest) (*srvpb.GetPriceReply, error) {
	respb := &srvpb.GetPriceReply{}
	c := strings.Replace(req.Name, " ", "", -1)
	search := fmt.Sprintf("https://api.scryfall.com/cards/search?q=!%v", c)
	response, err := http.Get(search)
	if err != nil {
		return nil, fmt.Errorf("Scryfall request failed: %s\n", err)
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		var sj ScryfallSearchJSON
		err = json.Unmarshal(data, &sj)
		if err != nil {
			fmt.Printf("data: %v\n", data)
			return nil, fmt.Errorf("Scryfall unmarshall err: %v\n", err)
		} else {
			if sj.Status == 404 {
				return nil, fmt.Errorf("404: %v", sj.Details)
			}
			if len(sj.Data) < 1 {
				return nil, fmt.Errorf("No data: %v", data)
			}
			usd, ok := sj.Data[0]["usd"]
			if !ok {
				return nil, fmt.Errorf("Data error: %v", sj)
			}
			respb.Price = usd.(string)
			return respb, nil
		}
	}
}

func (s *Service) ImportCollection(ctx context.Context, req *srvpb.ImportCollectionRequest) (*srvpb.ImportCollectionReply, error) {
	c, err := os.Open(req.FileName)
	if err != nil {
		return nil, err
	}
	r := csv.NewReader(bufio.NewReader(c))
	if _, ok := s.cols[req.Name]; !ok {
		s.cols[req.Name] = make([]*engine.Card, 0)
	}

	resp := &srvpb.ImportCollectionReply{}
	l, err := r.ReadAll()
	if err != nil {
		fmt.Printf("err: %v\n", l)
		return nil, err
	}
	added := 0
	unknown := make([]string, 0)
	for i, row := range l {
		if i == 0 {
			continue
		}
		name := string(row[0])
		card, err := s.searchExact(name)
		if err != nil {
			fmt.Printf("Couldn't find card: %v\n", name)
			unknown = append(unknown, name)
		} else {
			added++
			s.cols[req.Name] = append(s.cols[req.Name], card)
		}
	}
	resp.CardsAdded = int64(added)
	resp.UnknownCards = unknown
	return resp, nil
}

func (s *Service) GetStats(ctx context.Context, req *srvpb.GetStatsRequest) (*srvpb.GetStatsReply, error) {
	respb := &srvpb.GetStatsReply{}
	max, cmcs := colorCMC(s.cols[req.Collection])
	respb.Stats = colCMCprint(max, cmcs)
	respb.Stats = respb.Stats + "\n\n" + typeStats(s.cols[req.Collection])
	return respb, nil
}

func typeStats(cards []*engine.Card) string {
	stats := make(map[string]int)
	for _, c := range cards {
		for _, t := range c.Types {
			cnt, ok := stats[t]
			if !ok {
				cnt = 0
			}
			cnt++
			stats[t] = cnt
		}
	}
	out := "Type:\tCount\n"
	for t, c := range stats {
		out += fmt.Sprintf("%v:\t%v\n", t, c)
	}
	return out
}

type colorStats struct {
	color string
	cmcs map[int]int
	count int
}

func colorCMC(cards []*engine.Card) (int, []colorStats) {
	stats := make(map[string]colorStats)
	max := 0
	for _, c := range cards {
		col := color(c.Colors)
		s, ok := stats[col]
		if !ok {
			s = colorStats{color: col, cmcs: make(map[int]int), count: 0}
		}
		cmc := c.Cmc
		if cmc > max {
			max = cmc
		}
		cnt := s.cmcs[cmc]
		s.cmcs[cmc] = cnt + 1
		s.count++
		stats[col] = s
	}
	sts := make([]colorStats, 0)
	for _, v := range stats {
		sts = append(sts, v)
	}
	return max, sts
}

type byColor []colorStats

func (b byColor) Len() int { return len(b) }
func (b byColor) Less(i, j int) bool {
	// WUBRGC
	// TODO eww
	val := map[string]int{
		"W": 1,
		"U": 2,
		"B": 3,
		"R": 4,
		"G": 5,
		"WU": 6,
		"WB": 7,
		"WR": 8,
		"WG": 9,
		"UB": 10,
		"UR": 11,
		"UG": 12,
		"BR": 13,
		"BG": 14,
		"RG": 15,
		"WUB": 16,
		"WUR": 17,
		"WBR": 18,
		"WBG": 19,
		"WRG": 20,
		"UBR": 21,
		"URG": 22,
		"BRG": 23,
		"WUBR": 24,
		"WBRG": 25,
		"UBRG": 26,
		"WUBRG": 27,
	}
	return val[b[i].color] < val[b[j].color]
}
func (b byColor) Swap(i, j int) { b[i], b[j] = b[j], b[i] }

func colCMCprint(max int, colSts byColor) string {
	sort.Sort(colSts)
	out := "color\tcount\t"
	for i := 0; i <= max; i++ {
		out += fmt.Sprintf("%v\t", i)
	}
	out += "\n"
	for _, col := range colSts {
		o := fmt.Sprintf("%v\t%v\t", col.color, col.count)
		for i := 0; i <= max; i++ {
			o += fmt.Sprintf("%v\t", col.cmcs[i])
		}
		out += o + "\n"
	}
	return out
}
