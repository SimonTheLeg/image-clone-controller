package controller

import (
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// We make use of a customer predicate for two reasons:
// 1. We do not want to reconcile on delete events
// 2. We do not want to reconcile if the namespace is on the ignore list
func contrPredicate(igns []string) predicate.Predicate {
	contains := func(l []string, s string) bool {
		for _, v := range l {
			if v == s {
				return true
			}
		}
		return false
	}

	return predicate.Funcs{
		DeleteFunc: func(de event.DeleteEvent) bool {
			return false
		},
		CreateFunc: func(ce event.CreateEvent) bool {
			return !contains(igns, ce.Object.GetNamespace())
		},
		UpdateFunc: func(ue event.UpdateEvent) bool {
			return !contains(igns, ue.ObjectNew.GetNamespace())
		},
		GenericFunc: func(ge event.GenericEvent) bool {
			return !contains(igns, ge.Object.GetNamespace())
		},
	}
}
