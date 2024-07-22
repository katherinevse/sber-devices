package test_runner

import (
	"golang.org/x/net/html"
	"io"
	"net/url"
)

// ParseFormData разбирает данные формы из HTML.
func ParseFormData(body io.Reader) (url.Values, error) {
	values := url.Values{}
	tokenizer := html.NewTokenizer(body)
	radioValues := make(map[string]string)

	for {
		tokenType := tokenizer.Next()
		switch tokenType {
		case html.ErrorToken:
			return handleErrorToken(tokenizer, radioValues, values)
		case html.StartTagToken:
			handleStartTagToken(tokenizer, values, radioValues)
		}
	}
}

func handleErrorToken(tokenizer *html.Tokenizer, radioValues map[string]string, values url.Values) (url.Values, error) {
	err := tokenizer.Err()
	if err == io.EOF {
		for name, value := range radioValues {
			values.Set(name, value)
		}
		return values, nil
	}
	return nil, err
}

func handleStartTagToken(tokenizer *html.Tokenizer, values url.Values, radioValues map[string]string) {
	token := tokenizer.Token()
	switch token.Data {
	case "input":
		handleInputTag(token, values, radioValues)
	case "select":
		handleSelectTag(tokenizer, token, values)
	}
}

func handleInputTag(token html.Token, values url.Values, radioValues map[string]string) {
	var name, inputType, value string
	for _, attr := range token.Attr {
		switch attr.Key {
		case "name":
			name = attr.Val
		case "type":
			inputType = attr.Val
		case "value":
			value = attr.Val
		}
	}

	switch inputType {
	case "text":
		values.Set(name, "test")
	case "radio":
		if currentLongest, exists := radioValues[name]; !exists || len(value) > len(currentLongest) {
			radioValues[name] = value
		}
	}
}

func handleSelectTag(tokenizer *html.Tokenizer, token html.Token, values url.Values) {
	var name string
	for _, attr := range token.Attr {
		if attr.Key == "name" {
			name = attr.Val
		}
	}
	if name != "" {
		longestValue := findLongestOptionValue(tokenizer)
		values.Set(name, longestValue)
	}
}

// findLongestOptionValue находит самое длинное значение option внутри select.
func findLongestOptionValue(tokenizer *html.Tokenizer) string {
	longestValue := ""
	for {
		tokenType := tokenizer.Next()
		if tokenType == html.EndTagToken {
			token := tokenizer.Token()
			if token.Data == "select" {
				break
			}
		}
		if tokenType == html.StartTagToken {
			token := tokenizer.Token()
			if token.Data == "option" {
				value := attrValue(token, "value")
				if len(value) > len(longestValue) {
					longestValue = value
				}
			}
		}
	}
	return longestValue
}

// attrValue возвращает значение атрибута из токена.
func attrValue(token html.Token, key string) string {
	for _, attr := range token.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}
