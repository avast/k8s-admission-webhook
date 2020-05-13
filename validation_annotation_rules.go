package main

import (
	"github.com/ghodss/yaml"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"fmt"
	"io/ioutil"
	"strings"
)

func getAnnotationRulesFromConfig(config *config) *AnnotationRules {
	if config.AnnotationRulesFile == "" {
		log.Debugf("No annotation rules file specified - everything is allowed")
		return &AnnotationRules{
			AnnotationRules: map[string]AnnotationRule{},
		}
	}
	dat, err := ioutil.ReadFile(config.AnnotationRulesFile)
	checkFatal(err)
	var rules AnnotationRules
	err = yaml.Unmarshal(dat, &rules)
	checkFatal(err)
	return &rules
}

func getAnnotationRulesForKind(kind string, annotationRules *AnnotationRules) *AnnotationRule {
	for annotationRuleKind, annotationRule := range annotationRules.AnnotationRules {
		if annotationRuleKind == kind {
			log.Tracef("Rule %v matched for kind: %s", annotationRule, kind)
			return &annotationRule
		}
	}
	log.Tracef("No annotation rule match for: %s", kind)
	return nil
}

func validateAnnotationsByRules(validation *objectValidation, objectMeta *metav1.ObjectMeta, kind string, annotationRules *AnnotationRules) {
	log.Tracef("Validating %s with meta: %v", kind, objectMeta)
	annotationRule := getAnnotationRulesForKind(kind, annotationRules)

	targetDesc := fmt.Sprintf("Object %s.%s: ", objectMeta.Name, objectMeta.Namespace)
	for annotation := range objectMeta.GetAnnotations() {
		if annotationRule != nil {
			log.Tracef("Validating %s against %v", targetDesc, annotationRule)
			switch annotationRule.Policy {
			case AllowAll:
				for _, exception := range annotationRule.Exceptions {
					if strings.HasPrefix(annotation, exception) {
						validation.Violations.add(
							validationViolation{
								targetDesc,
								fmt.Sprintf("%s matches %s in exception list (for Allow policy) of %s annoration rules", annotation, exception, kind),
							},
						)
					}
				}
			case DenyAll:
				anyMatch := false
				for _, exception := range annotationRule.Exceptions {
					if strings.HasPrefix(annotation, exception) {
						anyMatch = true
					}
				}
				if !anyMatch {
					validation.Violations.add(
						validationViolation{
							targetDesc,
							fmt.Sprintf("%s doesn't matches any of %s in exception list (for Deny policy) of %s annoration rules", annotation, annotationRule.Exceptions, kind),
						},
					)
				}
			}
		}
	}
}

type AnnotationRule_Policy string

const (
	DenyAll  AnnotationRule_Policy = "DenyAll"
	AllowAll AnnotationRule_Policy = "AllowAll"
)

type AnnotationRule struct {
	Policy     AnnotationRule_Policy `json:"policy"`
	Exceptions []string              `json:"exceptions"`
}

type AnnotationRules struct {
	AnnotationRules map[string]AnnotationRule `json:"annotationRules"`
}
