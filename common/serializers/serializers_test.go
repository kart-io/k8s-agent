package serializers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestStruct struct {
	ID   string
	Name string
	Age  int
	Tags []string
}

func TestJSONSerializer(t *testing.T) {
	serializer := NewJSONSerializer()

	t.Run("Name", func(t *testing.T) {
		assert.Equal(t, "json", serializer.Name())
	})

	t.Run("Marshal and Unmarshal", func(t *testing.T) {
		original := TestStruct{
			ID:   "123",
			Name: "Alice",
			Age:  30,
			Tags: []string{"admin", "developer"},
		}

		// Marshal
		data, err := serializer.Marshal(original)
		require.NoError(t, err)
		require.NotEmpty(t, data)

		// Unmarshal
		var result TestStruct
		err = serializer.Unmarshal(data, &result)
		require.NoError(t, err)

		assert.Equal(t, original.ID, result.ID)
		assert.Equal(t, original.Name, result.Name)
		assert.Equal(t, original.Age, result.Age)
		assert.Equal(t, original.Tags, result.Tags)
	})

	t.Run("Marshal nil", func(t *testing.T) {
		data, err := serializer.Marshal(nil)
		require.NoError(t, err)
		assert.Equal(t, []byte("null"), data)
	})

	t.Run("Unmarshal invalid data", func(t *testing.T) {
		var result TestStruct
		err := serializer.Unmarshal([]byte("invalid json"), &result)
		assert.Error(t, err)
	})
}

func TestMsgpackSerializer(t *testing.T) {
	serializer := NewMsgpackSerializer()

	t.Run("Name", func(t *testing.T) {
		assert.Equal(t, "msgpack", serializer.Name())
	})

	t.Run("Marshal and Unmarshal", func(t *testing.T) {
		original := TestStruct{
			ID:   "456",
			Name: "Bob",
			Age:  25,
			Tags: []string{"user", "viewer"},
		}

		// Marshal
		data, err := serializer.Marshal(original)
		require.NoError(t, err)
		require.NotEmpty(t, data)

		// Unmarshal
		var result TestStruct
		err = serializer.Unmarshal(data, &result)
		require.NoError(t, err)

		assert.Equal(t, original.ID, result.ID)
		assert.Equal(t, original.Name, result.Name)
		assert.Equal(t, original.Age, result.Age)
		assert.Equal(t, original.Tags, result.Tags)
	})

	t.Run("Marshal nil", func(t *testing.T) {
		data, err := serializer.Marshal(nil)
		require.NoError(t, err)
		require.NotEmpty(t, data)
	})

	t.Run("Unmarshal invalid data", func(t *testing.T) {
		var result TestStruct
		err := serializer.Unmarshal([]byte("invalid msgpack"), &result)
		assert.Error(t, err)
	})
}

func TestYAMLSerializer(t *testing.T) {
	serializer := NewYAMLSerializer()

	t.Run("Name", func(t *testing.T) {
		assert.Equal(t, "yaml", serializer.Name())
	})

	t.Run("Marshal and Unmarshal", func(t *testing.T) {
		original := TestStruct{
			ID:   "789",
			Name: "Charlie",
			Age:  28,
			Tags: []string{"moderator", "contributor"},
		}

		// Marshal
		data, err := serializer.Marshal(original)
		require.NoError(t, err)
		require.NotEmpty(t, data)

		// Unmarshal
		var result TestStruct
		err = serializer.Unmarshal(data, &result)
		require.NoError(t, err)

		assert.Equal(t, original.ID, result.ID)
		assert.Equal(t, original.Name, result.Name)
		assert.Equal(t, original.Age, result.Age)
		assert.Equal(t, original.Tags, result.Tags)
	})

	t.Run("Marshal nil", func(t *testing.T) {
		data, err := serializer.Marshal(nil)
		require.NoError(t, err)
		require.NotEmpty(t, data)
	})

	t.Run("Unmarshal invalid data", func(t *testing.T) {
		var result TestStruct
		err := serializer.Unmarshal([]byte("invalid: yaml: ["), &result)
		assert.Error(t, err)
	})

	t.Run("Human readable format", func(t *testing.T) {
		original := TestStruct{
			ID:   "test-123",
			Name: "Test User",
			Age:  30,
			Tags: []string{"tag1", "tag2"},
		}

		data, err := serializer.Marshal(original)
		require.NoError(t, err)

		// YAML should be human-readable
		yamlStr := string(data)
		assert.Contains(t, yamlStr, "id: test-123")
		assert.Contains(t, yamlStr, "name: Test User")
		assert.Contains(t, yamlStr, "age: 30")
	})
}

func TestSerializerComparison(t *testing.T) {
	jsonSer := NewJSONSerializer()
	msgpackSer := NewMsgpackSerializer()

	testData := TestStruct{
		ID:   "comparison-test",
		Name: "Performance Test",
		Age:  42,
		Tags: []string{"benchmark", "performance", "comparison"},
	}

	t.Run("Size Comparison", func(t *testing.T) {
		jsonData, err := jsonSer.Marshal(testData)
		require.NoError(t, err)

		msgpackData, err := msgpackSer.Marshal(testData)
		require.NoError(t, err)

		t.Logf("JSON size: %d bytes", len(jsonData))
		t.Logf("Msgpack size: %d bytes", len(msgpackData))
		t.Logf("Size reduction: %.2f%%", (1.0-float64(len(msgpackData))/float64(len(jsonData)))*100)

		// Msgpack should be smaller
		assert.Less(t, len(msgpackData), len(jsonData), "Msgpack should be more compact than JSON")
	})

	t.Run("Cross-Serializer Compatibility", func(t *testing.T) {
		// JSON serialized data
		jsonData, err := jsonSer.Marshal(testData)
		require.NoError(t, err)

		var jsonResult TestStruct
		err = jsonSer.Unmarshal(jsonData, &jsonResult)
		require.NoError(t, err)

		// Msgpack serialized data
		msgpackData, err := msgpackSer.Marshal(testData)
		require.NoError(t, err)

		var msgpackResult TestStruct
		err = msgpackSer.Unmarshal(msgpackData, &msgpackResult)
		require.NoError(t, err)

		// Both should produce same result
		assert.Equal(t, jsonResult, msgpackResult)
	})
}

func BenchmarkSerializers(b *testing.B) {
	testData := TestStruct{
		ID:   "benchmark-test",
		Name: "Performance Benchmark",
		Age:  35,
		Tags: []string{"tag1", "tag2", "tag3", "tag4", "tag5"},
	}

	b.Run("JSON_Marshal", func(b *testing.B) {
		serializer := NewJSONSerializer()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = serializer.Marshal(testData)
		}
	})

	b.Run("JSON_Unmarshal", func(b *testing.B) {
		serializer := NewJSONSerializer()
		data, _ := serializer.Marshal(testData)
		var result TestStruct
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = serializer.Unmarshal(data, &result)
		}
	})

	b.Run("Msgpack_Marshal", func(b *testing.B) {
		serializer := NewMsgpackSerializer()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = serializer.Marshal(testData)
		}
	})

	b.Run("Msgpack_Unmarshal", func(b *testing.B) {
		serializer := NewMsgpackSerializer()
		data, _ := serializer.Marshal(testData)
		var result TestStruct
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = serializer.Unmarshal(data, &result)
		}
	})

	b.Run("YAML_Marshal", func(b *testing.B) {
		serializer := NewYAMLSerializer()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = serializer.Marshal(testData)
		}
	})

	b.Run("YAML_Unmarshal", func(b *testing.B) {
		serializer := NewYAMLSerializer()
		data, _ := serializer.Marshal(testData)
		var result TestStruct
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = serializer.Unmarshal(data, &result)
		}
	})
}
