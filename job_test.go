package wpool

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

var errDefault = errors.New("wrong argument type")
var descriptor = Descriptor{
	ID:   "alpha",
	Type: "something fancy",
	Metadata: map[string]interface{}{
		"class":    "default",
		"priority": 1,
	},
}
var intSquare = func(ctx context.Context, args interface{}) (interface{}, error) {
	v, ok := args.(int)
	if !ok {
		return nil, errDefault
	}
	return v * v, nil
}

func TestJob_execute(t *testing.T) {
	ctx := context.TODO()

	type fields struct {
		Descriptor Descriptor
		Task       Task
		Args       interface{}
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name   string
		fields fields
		want   Result
	}{
		{
			name: "success",
			fields: fields{
				Descriptor: descriptor,
				Task:       TaskFunc(intSquare),
				Args:       10,
			},
			want: Result{
				Value:      100,
				Descriptor: descriptor,
			},
		},
		{
			name: "failure",
			fields: fields{
				Descriptor: descriptor,
				Task:       TaskFunc(intSquare),
				Args:       "5",
			},
			want: Result{
				Err:        errDefault,
				Descriptor: descriptor,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := Job{
				Descriptor: tt.fields.Descriptor,
				Task:       tt.fields.Task,
				Args:       tt.fields.Args,
			}
			if got := j.execute(ctx); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("execute() = %v, want %v", got, tt.want)
			}
		})
	}
}
