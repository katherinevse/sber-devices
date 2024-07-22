package form

import (
	"golang.org/x/net/html"
	"io"
	"net/url"
)

//TODO архитектуру проекта сделать

func ParseFormData(body io.Reader) (url.Values, error) {
	values := url.Values{}
	tokenizer := html.NewTokenizer(body)

	radioValues := make(map[string]string)

	for {
		tokenType := tokenizer.Next()
		switch tokenType {
		case html.ErrorToken:
			err := tokenizer.Err()
			if err == io.EOF {
				for name, value := range radioValues {
					values.Set(name, value)
				}
				return values, nil
			}
			return nil, err

		case html.StartTagToken:
			token := tokenizer.Token()
			switch token.Data {
			case "input":
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
			case "select":
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
		}
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
