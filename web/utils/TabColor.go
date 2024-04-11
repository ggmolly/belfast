package utils

import "html/template"

func TabColor(page, currentPage string) template.HTML {
	if page == currentPage {
		return template.HTML("text-primary")
	}
	return template.HTML("text-secondary")
}
