package controller

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

func TestContrPredicate(t *testing.T) {

	pred := contrPredicate([]string{"kube-system"})

	// Cases
	// Namespace is on list
	// Namespace is not on list

	tt := map[string]struct {
		Ns    string
		expDe bool
		expCr bool
		expUp bool
		expGe bool
	}{
		"namespace should be ignored": {
			Ns:    "kube-system",
			expDe: false,
			expCr: false,
			expUp: false,
			expGe: false,
		},
		"namespace should watch, except delete": {
			Ns:    "some-namespace",
			expDe: false,
			expCr: true,
			expUp: true,
			expGe: true,
		},
	}

	for name, tc := range tt {

		t.Run(name, func(t *testing.T) {
			eventDe := event.DeleteEvent{Object: &appsv1.Deployment{}}
			eventDe.Object.SetNamespace(tc.Ns)
			gotDe := pred.Delete(eventDe)
			if gotDe != tc.expDe {
				t.Errorf("DeleteEvent: not correct predicate - exp '%t' got '%t'", tc.expDe, gotDe)
			}

			eventCr := event.CreateEvent{Object: &appsv1.Deployment{}}
			eventCr.Object.SetNamespace(tc.Ns)
			gotCr := pred.Create(eventCr)
			if gotCr != tc.expCr {
				t.Errorf("CreateEvent: not correct predicate - exp '%t' got '%t'", tc.expCr, gotCr)
			}

			eventUp := event.UpdateEvent{ObjectNew: &appsv1.Deployment{}}
			eventUp.ObjectNew.SetNamespace(tc.Ns)
			gotUp := pred.Update(eventUp)
			if gotUp != tc.expUp {
				t.Errorf("UpdateEvent: not correct predicate - exp '%t' got '%t'", tc.expUp, gotUp)
			}

			eventGe := event.GenericEvent{Object: &appsv1.Deployment{}}
			eventGe.Object.SetNamespace(tc.Ns)
			gotGe := pred.Generic(eventGe)
			if gotGe != tc.expGe {
				t.Errorf("GeeateEvent: not correct predicate - exp '%t' got '%t'", tc.expGe, gotGe)

			}
		})

	}
}
