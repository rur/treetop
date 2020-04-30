package inline

import "testing"

func TestFormData_MarshalBase64(t *testing.T) {
	tests := []struct {
		name     string
		data     *FormData
		wantErr  bool
		wantSize int
	}{
		{
			name:     "nil case",
			data:     &FormData{},
			wantSize: 28,
			wantErr:  false,
		},
		{
			name: "basic",
			data: &FormData{
				FirstName:   "John",
				LastName:    "Doe",
				Email:       "john.doe@gmail.com",
				Country:     "Canada",
				Description: "test",
			},
			wantErr:  false,
			wantSize: 64,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotB64, err := tt.data.MarshalBase64()
			if (err != nil) != tt.wantErr {
				t.Errorf("FormData.MarshalBase64() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			got := &FormData{}
			err = got.UnmarshalBase64(gotB64)

			if (err != nil) != tt.wantErr {
				t.Errorf("FormData.UnmarshalBase64 error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(gotB64) != tt.wantSize {
				t.Errorf("Expecting compressed B64 bytes size to be %d, got %d", tt.wantSize, len(gotB64))
			}

			if got.FirstName != tt.data.FirstName {
				t.Errorf("Expecting first name %v, got %v", tt.data.FirstName, got.FirstName)
			}
			if got.LastName != tt.data.LastName {
				t.Errorf("Expecting last name %v, got %v", tt.data.LastName, got.LastName)
			}
			if got.Email != tt.data.Email {
				t.Errorf("Expecting email %v, got %v", tt.data.Email, got.Email)
			}
			if got.Country != tt.data.Country {
				t.Errorf("Expecting Country %v, got %v", tt.data.Country, got.Country)
			}
			if got.Description != tt.data.Description {
				t.Errorf("Expecting Description %v, got %v", tt.data.Description, got.Description)
			}
		})
	}
}
