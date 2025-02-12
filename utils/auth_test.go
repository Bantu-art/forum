package utils

import (
	"regexp"
	"testing"
)

func TestHashPassword(t *testing.T) {
	type args struct {
		password string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"Valid Password", args{"SecurePass123!"}, false},
		{"Empty Password", args{""}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := HashPassword(tt.args.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("HashPassword() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(got) == 0 {
				t.Errorf("HashPassword() returned empty hash")
			}
		})
	}
}

func TestCheckPasswordsHash(t *testing.T) {
	password := "SecurePass123!"
	hashed, _ := HashPassword(password)
	tests := []struct {
		name string
		args struct {
			password string
			hash     string
		}
		want bool
	}{
		{"Correct Password", struct{ password, hash string }{password, hashed}, true},
		{"Incorrect Password", struct{ password, hash string }{"WrongPass", hashed}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckPasswordsHash(tt.args.password, tt.args.hash); got != tt.want {
				t.Errorf("CheckPasswordsHash() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  bool
	}{
		{"Valid Email", "test@example.com", true},
		{"Invalid Email - No @", "testexample.com", false},
		{"Invalid Email - No domain", "test@.com", false},
		{"Invalid Email - Empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateEmail(tt.email); got != tt.want {
				t.Errorf("ValidateEmail() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		want     bool
	}{
		{"Valid Username", "JohnDoe", true},
		{"Too Short", "Jo", false},
		{"Too Long", "ThisIsAReallyLongUsernameThatExceedsThirtyCharacters", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateUsername(tt.username); got != tt.want {
				t.Errorf("ValidateUsername() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		want     bool
	}{
		{"Valid Password", "P@ssw0rd!", true},
		{"No Uppercase", "p@ssw0rd!", false},
		{"No Lowercase", "P@SSW0RD!", false},
		{"No Number", "P@ssword!", false},
		{"No Special Character", "Passw0rd", false},
		{"Too Short", "P@1!", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidatePassword(tt.password); got != tt.want {
				t.Errorf("ValidatePassword() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateId(t *testing.T) {
	validUUID := regexp.MustCompile(`^[a-f0-9-]{36}$`)
	t.Run("Valid UUID", func(t *testing.T) {
		got := GenerateId()
		if !validUUID.MatchString(got) {
			t.Errorf("GenerateId() = %v, not a valid UUID", got)
		}
	})
}
