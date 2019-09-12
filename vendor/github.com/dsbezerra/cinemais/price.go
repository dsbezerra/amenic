package cinemais

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/dsbezerra/cinemais/util"
	"github.com/gocolly/colly"
)

const (
	// Labels
	PricesLabel2D        = "Valores para projeção 2D"                            // 2D prices
	PricesLabel3D        = "Valores para projeção 3D"                            // 3D prices
	PricesLabelMagicDVIP = "Sala MAGIC D (Sessões em 2D e 3D - Poltrona VIP)"    // Magic D - VIP 2D/3D
	PricesLabelMagicD2D  = "Sala MAGIC D (Sessões em 2D - Poltrona Tradicional)" // Magic D - 2D
	PricesLabelMagicD3D  = "Sala MAGIC D (Sessões em 3D - Poltrona Tradicional)" // Magic D - 3D
)

// Price ...
type Price struct {
	Label             string         `json:"label"`
	Attributes        []string       `json:"attributes"`
	Weekdays          []time.Weekday `json:"weekdays"`
	Full              float32        `json:"full"`
	Half              float32        `json:"half"`
	ExceptHolidays    bool           `json:"except_holidays"`
	ExceptPreviews    bool           `json:"except_previews"`
	IncludingHolidays bool           `json:"including_holidays"`
	IncludingPreviews bool           `json:"including_previews"`
}

// Prices ...
type Prices []Price

// GetPrices ...
func GetPrices(ID string) (Prices, error) {
	return New(CinemaCode(ID)).GetPrices()
}

// GetPrices ...
func (c *Cinemais) GetPrices() (Prices, error) {
	var result Prices
	var err error

	c.collector.OnHTML("body", func(e *colly.HTMLElement) {
		result, err = parsePrices(strings.TrimSpace(e.Text))
	})

	c.collector.OnError(func(r *colly.Response, e error) {
		err = e
	})

	url := fmt.Sprintf("%s/programacao/ingresso_velox.php?cc=%s", c.BaseURL.String(), c.CinemaCode)
	c.collector.Visit(url)
	return result, err
}

func parsePrices(text string) (Prices, error) {
	if text == "" {
		return nil, ErrUnexpectedStructure
	}

	parsing := ""

	result := make([]Price, 0)
	scanner := bufio.NewScanner(strings.NewReader(text))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "TABELA DE PREÇOS" || line == "" {
			continue
		}

		if strings.Contains(line, PricesLabel2D) {
			parsing = PricesLabel2D
			continue
		} else if strings.Contains(line, PricesLabel3D) {
			parsing = PricesLabel3D
			continue
		} else if strings.Contains(line, PricesLabelMagicDVIP) {
			parsing = PricesLabelMagicDVIP
			continue
		} else if strings.Contains(line, PricesLabelMagicD2D) {
			parsing = PricesLabelMagicD2D
			continue
		} else if strings.Contains(line, PricesLabelMagicD3D) {
			parsing = PricesLabelMagicD3D
			continue
		}

		var attrs []string
		var label string

		switch parsing {
		case PricesLabel2D:
			attrs = []string{"2D"}
			label = "Projeção 2D"
		case PricesLabel3D:
			attrs = []string{"3D"}
			label = "Projeção 3D"
		case PricesLabelMagicD2D:
			attrs = []string{"2D", "Magic D", "Poltrona Tradicional"}
			label = "Magic D - Projeção 2D - Poltrona Tradicional"
		case PricesLabelMagicD3D:
			attrs = []string{"3D", "Magic D", "Poltrona Tradicional"}
			label = "Magic D - Projeção 3D - Poltrona Tradicional"
		case PricesLabelMagicDVIP:
			attrs = []string{"2D", "3D", "Magic D", "Poltrona VIP"}
			label = "Magic D - Projeção 2D/3D - Poltrona VIP"
		}

		if label == "" || len(attrs) == 0 {
			// TODO: Logging
		} else {
			price, err := parsePrice(line)
			if err == nil {
				price.Label = label
				price.Attributes = attrs
				result = append(result, *price)
			} else {
				// TODO(diego): Logging
			}
		}
	}

	return result, nil
}

func parsePrice(text string) (*Price, error) {
	weekdays, prices := util.BreakByToken(text, ':')
	if weekdays == "" || prices == "" {
		return nil, ErrUnexpectedStructure
	}

	var price Price

	for {
		weekday, remainder := util.BreakBySpaces(weekdays)
		weekdays = remainder
		if weekday != "" {
			day, ok := util.ParseDay(weekday)
			if ok {
				exists := false
				for _, w := range price.Weekdays {
					if w == day {
						exists = true
						break
					}
				}
				if !exists {
					price.Weekdays = append(price.Weekdays, day)
				}
			} else {
				if weekday == "feriados" {
					price.IncludingHolidays = true
				} else if weekday == "pré-estreias" {
					price.IncludingPreviews = true
				}
			}

			if remainder == "" {
				break
			}

			// Parse except
			if remainder[0] == '(' {
				util.StringAdvance(&remainder, 1)
				cond, days := util.BreakBySpaces(remainder)
				if cond == "exceto" {
					for {
						d, remainder := util.BreakBySpaces(days)
						if d == "feriados" {
							price.ExceptHolidays = true
						} else if d == "pré-estreias" {
							price.ExceptPreviews = true
						}
						days = remainder
						if remainder == "" {
							break
						}
					}
				}
				break
			}
		}
	}

	// Get value from text
	valueText := strings.Builder{}
	i := 0
	consumedAny := false
	for i < len(prices) {
		r := rune(prices[i])
		if unicode.IsDigit(r) {
			valueText.WriteRune(r)
			consumedAny = true
		} else if r == ',' {
			valueText.WriteString(".")
			consumedAny = true
		} else if unicode.IsSpace(r) && consumedAny {
			break
		}
		i++
	}

	value, err := strconv.ParseFloat(valueText.String(), 32)
	if err != nil {
		return nil, ErrUnexpectedStructure
	}

	price.Full = float32(value)
	price.Half = price.Full / 2.0

	return &price, nil
}
