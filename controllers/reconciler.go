package controllers

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type Reconciler struct{}

func (r *Reconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	fmt.Println("Reconciling something!")
	return reconcile.Result{}, nil
}
