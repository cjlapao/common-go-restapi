package restapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewHttpListener_CreatesHttpListenerWithDefaultParams(t *testing.T) {
	listener := NewHttpListener()

	assert.NotNilf(t, listener, "Listener should not be empty")
	assert.NotNilf(t, listener.Options, "Listener Options should not be empty")
	assert.NotNilf(t, listener.Logger, "Listener Logger should not be empty")
	assert.NotNilf(t, listener.Router, "Listener Router should not be empty")
	assert.NotNilf(t, listener.Context, "Listener Context should not be empty")
	assert.Lenf(t, listener.Controllers, 0, "Controllers should be empty")
	assert.Equalf(t, "5000", listener.Options.HttpPort, "Http port should be 5000")
	assert.Equalf(t, "5001", listener.Options.TLSPort, "Https port should be 5001")
	assert.Equalf(t, "users", listener.Options.DatabaseName, "Authentication database name should be users")
	assert.Falsef(t, listener.Options.UseAuthBackend, "Backend Should be false")
}

func Test_joinUrl(t *testing.T) {
	type args struct {
		element []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "no elements",
			args: args{},
			want: "/",
		},
		{
			name: "mixed elements",
			args: args{
				element: []string{
					"/foo",
					"bar",
				},
			},
			want: "/foo/bar",
		},
		{
			name: "mixed elements suffix",
			args: args{
				element: []string{
					"/foo",
					"bar/",
				},
			},
			want: "/foo/bar",
		},
		{
			name: "mixed elements prefix",
			args: args{
				element: []string{
					"/foo/",
					"bar",
				},
			},
			want: "/foo/bar",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := joinUrl(tt.args.element...); got != tt.want {
				t.Errorf("joinUrl() = %v, want %v", got, tt.want)
			}
		})
	}
}
