package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type SpinRequest struct {
	Bet         float64        `json:"bet"`
	FreeSpins   int            `json:"freeSpins"`
	StickyWilds map[string]int `json:"stickyWilds"`
}

type SpinResult struct {
	Reels        [][]string     `json:"reels"`
	WinAmount    float64        `json:"win"`
	WinningLines [][]int        `json:"winningLines"`
	FreeSpins    int            `json:"freeSpins"`
	StickyWilds  map[string]int `json:"stickyWilds"`
}

var payouts = map[string][]float64{
	"LION":    {0, 0.5, 1.5, 7.5},
	"TIGER":   {0, 0.35, 1.0, 5.0},
	"PANTHER": {0, 0.25, 0.6, 3.0},
	"CAT":     {0, 0.2, 0.4, 2.0},
	"STEAK":   {0, 0.12, 0.25, 1.5},
	"BONE":    {0, 0.08, 0.2, 1.0},
	"A":       {0, 0.05, 0.1, 0.5},
	"K":       {0, 0.05, 0.1, 0.5},
	"Q":       {0, 0.02, 0.05, 0.25},
	"J":       {0, 0.02, 0.05, 0.25},
	"TEN":     {0, 0.02, 0.05, 0.25},
}

var payLines = [][]int{
	{1, 1, 1, 1, 1}, {0, 0, 0, 0, 0}, {2, 2, 2, 2, 2},
	{0, 1, 2, 1, 0}, {2, 1, 0, 1, 2}, {0, 0, 1, 0, 0},
	{2, 2, 1, 2, 2}, {1, 0, 0, 0, 1}, {1, 2, 2, 2, 1},
	{0, 1, 1, 1, 0}, {2, 1, 1, 1, 2}, {1, 1, 0, 1, 1},
	{1, 1, 2, 1, 1}, {0, 1, 0, 1, 0}, {2, 1, 2, 1, 2},
	{1, 0, 1, 0, 1}, {1, 2, 1, 2, 1}, {0, 1, 2, 2, 2},
	{2, 1, 0, 0, 0}, {0, 1, 2, 1, 0},
}

var reelStrips = [][]string{
	{
		"LION", "TEN", "CAT", "K", "STEAK", "Q", "TIGER", "A", "PANTHER", "J",
		"BONE", "Q", "LION", "K", "TEN", "STEAK", "J", "CAT", "Q", "TIGER",
		"BONE", "A", "PANTHER", "K", "J", "LION", "A", "K", "TEN", "PANTHER",
		"STEAK", "Q", "TIGER", "J", "CAT", "BONE", "Q", "LION", "K", "STEAK",
		"PANTHER", "A", "TIGER", "J", "TEN", "CAT", "Q", "STEAK", "LION", "BONE",
	},
	{
		"TIGER", "Q", "LION", "BONE", "K", "A", "PANTHER", "TEN", "STEAK", "J",
		"CAT", "K", "A", "TIGER", "STEAK", "Q", "PANTHER", "J", "LION", "BONE",
		"K", "CAT", "TEN", "Q", "A", "PANTHER", "K", "STEAK", "LION", "TEN",
		"TIGER", "J", "Q", "BONE", "STEAK", "K", "LION", "J", "CAT", "A",
		"Q", "PANTHER", "TEN", "BONE", "K", "TIGER", "STEAK", "Q", "LION", "A",
	},
	{
		"PANTHER", "BONE", "K", "LION", "STEAK", "Q", "TIGER", "A", "CAT", "J",
		"STEAK", "Q", "LION", "TEN", "PANTHER", "A", "BONE", "Q", "TIGER", "K",
		"J", "STEAK", "CAT", "K", "TEN", "LION", "BONE", "Q", "PANTHER", "A",
		"STEAK", "TIGER", "J", "Q", "CAT", "K", "A", "BONE", "STEAK", "TEN",
		"LION", "PANTHER", "Q", "TIGER", "J", "BONE", "CAT", "K", "STEAK", "Q",
	},
	{
		"CAT", "K", "STEAK", "J", "Q", "PANTHER", "LION", "A", "TIGER", "TEN",
		"STEAK", "K", "PANTHER", "BONE", "Q", "LION", "A", "TIGER", "J", "CAT",
		"Q", "K", "TEN", "STEAK", "PANTHER", "LION", "BONE", "A", "J", "Q",
		"TIGER", "CAT", "K", "STEAK", "TEN", "BONE", "Q", "LION", "A", "PANTHER",
		"K", "TIGER", "J", "CAT", "Q", "BONE", "LION", "STEAK", "TEN", "K",
	},
	{
		"TEN", "STEAK", "LION", "A", "TIGER", "K", "Q", "PANTHER", "STEAK", "J",
		"BONE", "CAT", "A", "TIGER", "TEN", "K", "PANTHER", "Q", "LION", "STEAK",
		"BONE", "J", "K", "TEN", "PANTHER", "A", "TIGER", "LION", "Q", "STEAK",
		"BONE", "J", "CAT", "TEN", "K", "PANTHER", "Q", "TIGER", "LION", "A",
		"STEAK", "J", "BONE", "K", "CAT", "Q", "TEN", "TIGER", "LION", "BONE",
	},
}

func key(i, j int) string {
	return strconv.Itoa(i) + "-" + strconv.Itoa(j)
}

func generateReels(stickyWilds map[string]int, freeSpins int) [][]string {
	reels := make([][]string, 5)
	for i := range reels {
		strip := reelStrips[i]
		start := rand.Intn(len(strip))
		reels[i] = make([]string, 3)
		for j := 0; j < 3; j++ {
			k := key(i, j)
			if stickyWilds != nil && stickyWilds[k] != 0 {
				if stickyWilds[k] == 2 {
					reels[i][j] = "WILDX2"
				} else {
					reels[i][j] = "WILDX3"
				}
			} else {
				symbol := strip[(start+j)%len(strip)]
				if freeSpins > 0 && symbol == "BONUS" {
					symbol = "TEN"
				}
				reels[i][j] = symbol
			}
		}
	}
	return reels
}

func calculateWin(reels [][]string, bet float64) (float64, [][]int) {
	totalWin := 0.0
	winningLines := [][]int{}

	for _, line := range payLines {
		streak := 0
		baseSymbol := ""
		multiplier := 1

		for i := 0; i < 5; i++ {
			row := line[i]
			symbol := reels[i][row]

			if symbol == "BONUS" {
				break
			}

			if symbol == "WILDX2" || symbol == "WILDX3" {
				if i == 0 || baseSymbol == "" {
					break
				}
				streak++
				if symbol == "WILDX2" {
					multiplier *= 2
				} else {
					multiplier *= 3
				}
			} else {
				if baseSymbol == "" {
					baseSymbol = symbol
					streak = 1
				} else if symbol == baseSymbol {
					streak++
				} else {
					break
				}
			}
		}

		if streak >= 3 && baseSymbol != "" {
			payoutsForSymbol, exists := payouts[baseSymbol]
			if !exists {
				continue
			}
			if streak >= len(payoutsForSymbol) {
				streak = len(payoutsForSymbol) - 1
			}
			win := payoutsForSymbol[streak] * float64(multiplier) * bet
			totalWin += win
			winningLines = append(winningLines, line)
		}
	}

	return totalWin, winningLines
}

func main() {
	rand.Seed(time.Now().UnixNano())
	r := gin.Default()

	r.POST("/spin", func(c *gin.Context) {
		var req SpinRequest
		if err := c.BindJSON(&req); err != nil || req.Bet < 0.2 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Aposta inválida. Mínimo: €0.20"})
			return
		}

		reels := generateReels(req.StickyWilds, req.FreeSpins)
		win, winningLines := calculateWin(reels, req.Bet)

		bonusCount := 0
		for _, pos := range []int{0, 2, 4} {
			for _, symbol := range reels[pos] {
				if symbol == "BONUS" {
					bonusCount++
				}
			}
		}

		freeSpins := req.FreeSpins
		stickyWilds := req.StickyWilds
		if stickyWilds == nil {
			stickyWilds = make(map[string]int)
		}

		if bonusCount == 3 {
			freeSpins = 12
			win += 5.0 * req.Bet
		} else if freeSpins > 0 {
			for i := 1; i <= 3; i++ {
				for j := 0; j < 3; j++ {
					symbol := reels[i][j]
					if symbol == "WILDX2" {
						stickyWilds[key(i, j)] = 2
					} else if symbol == "WILDX3" {
						stickyWilds[key(i, j)] = 3
					}
				}
			}
			freeSpins--
		}

		c.JSON(http.StatusOK, SpinResult{
			Reels:        reels,
			WinAmount:    win,
			WinningLines: winningLines,
			FreeSpins:    freeSpins,
			StickyWilds:  stickyWilds,
		})

		fmt.Println("Value of reels:", reels)
	})

	r.Run(":8080")
}
