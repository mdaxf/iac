package funcs

import (
	"testing"
	"time"

	"github.com/mdaxf/iac/engine/types"
)

// TestTypeConverter_ConvertToInt demonstrates table-driven testing
func TestTypeConverter_ConvertToInt(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		want    int
		wantErr bool
	}{
		{
			name:    "string to int",
			input:   "123",
			want:    123,
			wantErr: false,
		},
		{
			name:    "float to int",
			input:   123.45,
			want:    123,
			wantErr: false,
		},
		{
			name:    "bool true to int",
			input:   true,
			want:    1,
			wantErr: false,
		},
		{
			name:    "bool false to int",
			input:   false,
			want:    0,
			wantErr: false,
		},
		{
			name:    "int to int",
			input:   456,
			want:    456,
			wantErr: false,
		},
		{
			name:    "invalid string",
			input:   "abc",
			want:    0,
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			want:    0,
			wantErr: false,
		},
	}

	tc := &TypeConverter{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tc.ConvertToInt(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertToInt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("ConvertToInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestTypeConverter_ConvertToFloat demonstrates testing float conversions
func TestTypeConverter_ConvertToFloat(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		want    float64
		wantErr bool
	}{
		{
			name:    "string to float",
			input:   "123.45",
			want:    123.45,
			wantErr: false,
		},
		{
			name:    "int to float",
			input:   123,
			want:    123.0,
			wantErr: false,
		},
		{
			name:    "bool true to float",
			input:   true,
			want:    1.0,
			wantErr: false,
		},
		{
			name:    "empty string",
			input:   "",
			want:    0.0,
			wantErr: false,
		},
	}

	tc := &TypeConverter{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tc.ConvertToFloat(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertToFloat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("ConvertToFloat() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestTypeConverter_ConvertToBool demonstrates testing bool conversions
func TestTypeConverter_ConvertToBool(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		want    bool
		wantErr bool
	}{
		{
			name:    "string true",
			input:   "true",
			want:    true,
			wantErr: false,
		},
		{
			name:    "string false",
			input:   "false",
			want:    false,
			wantErr: false,
		},
		{
			name:    "string 1",
			input:   "1",
			want:    true,
			wantErr: false,
		},
		{
			name:    "string 0",
			input:   "0",
			want:    false,
			wantErr: false,
		},
		{
			name:    "string yes",
			input:   "yes",
			want:    true,
			wantErr: false,
		},
		{
			name:    "string no",
			input:   "no",
			want:    false,
			wantErr: false,
		},
		{
			name:    "bool true",
			input:   true,
			want:    true,
			wantErr: false,
		},
		{
			name:    "int 1",
			input:   1,
			want:    true,
			wantErr: false,
		},
		{
			name:    "int 0",
			input:   0,
			want:    false,
			wantErr: false,
		},
	}

	tc := &TypeConverter{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tc.ConvertToBool(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertToBool() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("ConvertToBool() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestOutputBuilder demonstrates testing output builder
func TestOutputBuilder(t *testing.T) {
	t.Run("Set simple values", func(t *testing.T) {
		builder := NewOutputBuilder()

		builder.Set("name", "John").
			Set("age", 30).
			Set("active", true)

		outputs, err := builder.Build()

		if err != nil {
			t.Errorf("Build() unexpected error: %v", err)
		}

		if outputs["name"] != "John" {
			t.Errorf("Expected name to be John, got %v", outputs["name"])
		}

		if outputs["age"] != 30 {
			t.Errorf("Expected age to be 30, got %v", outputs["age"])
		}

		if outputs["active"] != true {
			t.Errorf("Expected active to be true, got %v", outputs["active"])
		}
	})

	t.Run("SetInt with conversion", func(t *testing.T) {
		builder := NewOutputBuilder()

		builder.SetInt("count", "123")

		outputs, err := builder.Build()

		if err != nil {
			t.Errorf("Build() unexpected error: %v", err)
		}

		if outputs["count"] != 123 {
			t.Errorf("Expected count to be 123, got %v", outputs["count"])
		}
	})
}

// TestSessionHelper demonstrates testing session operations
func TestSessionHelper(t *testing.T) {
	systemSession := map[string]interface{}{
		"UserNo":   "user123",
		"ClientID": "client456",
	}

	userSession := map[string]interface{}{
		"Language": "en",
		"Theme":    "dark",
	}

	helper := NewSessionHelper(systemSession, userSession)

	t.Run("GetSystemString", func(t *testing.T) {
		userNo := helper.GetSystemString("UserNo", "default")
		if userNo != "user123" {
			t.Errorf("Expected user123, got %s", userNo)
		}

		missing := helper.GetSystemString("Missing", "default")
		if missing != "default" {
			t.Errorf("Expected default, got %s", missing)
		}
	})

	t.Run("GetUserString", func(t *testing.T) {
		lang := helper.GetUserString("Language", "default")
		if lang != "en" {
			t.Errorf("Expected en, got %s", lang)
		}
	})
}

// TestSliceHelper demonstrates slice utility testing
func TestSliceHelper(t *testing.T) {
	helper := &SliceHelper{}

	t.Run("ContainsString", func(t *testing.T) {
		slice := []string{"apple", "banana", "cherry"}

		if !helper.ContainsString(slice, "banana") {
			t.Error("Expected to find banana")
		}

		if helper.ContainsString(slice, "grape") {
			t.Error("Expected not to find grape")
		}
	})

	t.Run("UniqueStrings", func(t *testing.T) {
		slice := []string{"apple", "banana", "apple", "cherry", "banana"}
		unique := helper.UniqueStrings(slice)

		if len(unique) != 3 {
			t.Errorf("Expected 3 unique strings, got %d", len(unique))
		}
	})

	t.Run("FilterStrings", func(t *testing.T) {
		slice := []string{"apple", "banana", "apricot", "cherry"}

		filtered := helper.FilterStrings(slice, func(s string) bool {
			return len(s) > 0 && s[0] == 'a'
		})

		if len(filtered) != 2 {
			t.Errorf("Expected 2 filtered strings, got %d", len(filtered))
		}
	})
}

// TestStringHelper demonstrates string utility testing
func TestStringHelper(t *testing.T) {
	helper := &StringHelper{}

	t.Run("IsEmpty", func(t *testing.T) {
		if !helper.IsEmpty("") {
			t.Error("Expected empty string to be empty")
		}

		if !helper.IsEmpty("   ") {
			t.Error("Expected whitespace to be empty")
		}

		if helper.IsEmpty("test") {
			t.Error("Expected 'test' not to be empty")
		}
	})

	t.Run("DefaultIfEmpty", func(t *testing.T) {
		result := helper.DefaultIfEmpty("", "default")
		if result != "default" {
			t.Errorf("Expected default, got %s", result)
		}

		result = helper.DefaultIfEmpty("value", "default")
		if result != "value" {
			t.Errorf("Expected value, got %s", result)
		}
	})

	t.Run("Truncate", func(t *testing.T) {
		long := "This is a very long string"
		truncated := helper.Truncate(long, 10)

		if truncated != "This is a ..." {
			t.Errorf("Expected truncation, got %s", truncated)
		}

		short := "Short"
		notTruncated := helper.Truncate(short, 10)

		if notTruncated != "Short" {
			t.Errorf("Expected no truncation, got %s", notTruncated)
		}
	})
}

// BenchmarkTypeConverter_ConvertToInt demonstrates benchmarking
func BenchmarkTypeConverter_ConvertToInt(b *testing.B) {
	tc := &TypeConverter{}

	b.Run("string", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			tc.ConvertToInt("123")
		}
	})

	b.Run("int", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			tc.ConvertToInt(123)
		}
	})

	b.Run("float", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			tc.ConvertToInt(123.45)
		}
	})
}

// BenchmarkOutputBuilder demonstrates output builder benchmarking
func BenchmarkOutputBuilder(b *testing.B) {
	b.Run("Build with 5 fields", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			builder := NewOutputBuilder()
			builder.Set("field1", "value1").
				Set("field2", 123).
				Set("field3", true).
				Set("field4", 45.67).
				Set("field5", time.Now())
			builder.Build()
		}
	})
}

// Example_TypeConverter demonstrates example usage
func Example_TypeConverter() {
	tc := &TypeConverter{}

	// Convert string to int
	intVal, _ := tc.ConvertToInt("123")
	println(intVal) // 123

	// Convert string to bool
	boolVal, _ := tc.ConvertToBool("true")
	println(boolVal) // true

	// Convert interface to string
	strVal := tc.ConvertToString(123.45)
	println(strVal) // "123.45"
}

// Example_OutputBuilder demonstrates output builder usage
func Example_OutputBuilder() {
	builder := NewOutputBuilder()

	builder.SetString("name", "John Doe").
		SetInt("age", "30").
		SetBool("active", "true")

	outputs, _ := builder.Build()
	println(outputs["name"]) // "John Doe"
}
