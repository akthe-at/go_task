package data

import (
	"testing"
)

func TestStringToPriorityType(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    PriorityType
		wantErr bool
	}{
		{
			name:    "invalid blank priority",
			args:    args{s: ""},
			want:    "",
			wantErr: true,
		},
		{
			name:    "valid low priority",
			args:    args{s: "low"},
			want:    PriorityTypeLow,
			wantErr: false,
		},
		{
			name:    "valid medium priority",
			args:    args{s: "medium"},
			want:    PriorityTypeMedium,
			wantErr: false,
		},
		{
			name:    "valid high priority",
			args:    args{s: "high"},
			want:    PriorityTypeHigh,
			wantErr: false,
		},
		{
			name:    "valid urgent priority",
			args:    args{s: "urgent"},
			want:    PriorityTypeUrgent,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := StringToPriorityType(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("StringToPriorityType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("StringToPriorityType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStringToStatusType(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    StatusType
		wantErr bool
	}{
		{
			name:    "valid todo status",
			args:    args{s: "todo"},
			want:    StatusToDo,
			wantErr: false,
		},
		{
			name:    "valid planning status",
			args:    args{s: "planning"},
			want:    StatusPlanning,
			wantErr: false,
		},
		{
			name:    "valid doing status",
			args:    args{s: "doing"},
			want:    StatusDoing,
			wantErr: false,
		},
		{
			name:    "valid done status",
			args:    args{s: "done"},
			want:    StatusDone,
			wantErr: false,
		},
		{
			name:    "Invalid status",
			args:    args{s: ""},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := StringToStatusType(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("StringToStatusType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("StringToStatusType() = %v, want %v", got, tt.want)
			}
		})
	}
}
