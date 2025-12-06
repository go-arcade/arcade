package parallel

import (
	"context"
	"reflect"
	"testing"
	"time"
)

func TestGo(t *testing.T) {
	type args struct {
		ctx  context.Context
		fn   func(ctx context.Context) (interface{}, error)
		opts []RunOption
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr interface{}
	}{
		{
			args: args{
				ctx: context.Background(),
				fn: func(ctx context.Context) (interface{}, error) {
					return 1, nil
				},
			},
			want:    1,
			wantErr: nil,
		},
		{
			args: args{
				ctx: context.Background(),
				fn: func(ctx context.Context) (interface{}, error) {
					return nil, context.Canceled
				},
			},
			want:    nil,
			wantErr: context.Canceled,
		},
		{
			args: args{
				ctx: context.Background(),
				fn: func(ctx context.Context) (any, error) {
					<-ctx.Done()
					return nil, ctx.Err()
				},
				opts: []RunOption{WithTimeout(1 * time.Second)},
			},
			want:    nil,
			wantErr: context.DeadlineExceeded,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			future := Go(tt.args.ctx, tt.args.fn, tt.args.opts...)

			if data, err := future.Get(); !reflect.DeepEqual(data, tt.want) || !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("Go() And Get() = (%v, %v), want (%v, %v)", data, err, tt.want, tt.wantErr)
			}
		})
	}
}
