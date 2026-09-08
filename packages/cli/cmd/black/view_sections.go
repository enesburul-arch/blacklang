package main

func indexComponents(components []ComponentDecl) map[string]ComponentDecl {
	index := map[string]ComponentDecl{}
	for _, component := range components {
		if _, exists := index[component.Name]; exists {
			continue
		}
		index[component.Name] = component
	}
	return index
}

func pageViewSectionSupported(page PageDecl, sectionName string) bool {
	if supportedViewSections[sectionName] {
		return true
	}
	_, ok := pageViewComponentSection(page, sectionName)
	return ok
}

func pageViewComponentSections(page PageDecl) []ViewSectionDecl {
	if page.View == nil {
		return nil
	}
	sections := []ViewSectionDecl{}
	for _, section := range page.View.Sections {
		if section.Component != "" {
			sections = append(sections, section)
		}
	}
	return sections
}

func pageViewComponentSection(page PageDecl, sectionName string) (ViewSectionDecl, bool) {
	if page.View == nil {
		return ViewSectionDecl{}, false
	}
	for _, section := range page.View.Sections {
		if section.Name == sectionName && section.Component != "" {
			return section, true
		}
	}
	return ViewSectionDecl{}, false
}
