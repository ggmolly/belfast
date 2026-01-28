package rng

import (
	"reflect"
	"testing"
)

func TestNewLockedRand(t *testing.T) {
	r := NewLockedRand()

	if r.r == nil {
		t.Fatalf("expected internal rand to be initialized")
	}

	value := r.Uint32()

	if value > ^uint32(0) {
		t.Fatalf("expected Uint32 to return valid value, got %d", value)
	}
}

func TestNewLockedRandFromSeed(t *testing.T) {
	seed := uint64(12345)
	r1 := NewLockedRandFromSeed(seed)
	r2 := NewLockedRandFromSeed(seed)

	if r1.r == nil || r2.r == nil {
		t.Fatalf("expected internal rand to be initialized")
	}

	v1 := r1.Uint32()
	v2 := r2.Uint32()

	if v1 != v2 {
		t.Fatalf("expected same seed to produce same value, got %d and %d", v1, v2)
	}
}

func TestUint32(t *testing.T) {
	r := NewLockedRand()

	values := make(map[uint32]bool)
	for i := 0; i < 100; i++ {
		value := r.Uint32()
		values[value] = true
		if value > ^uint32(0) {
			t.Fatalf("expected Uint32 to return value in valid range, got %d", value)
		}
	}

	if len(values) < 50 {
		t.Fatalf("expected Uint32 to produce different values, got only %d unique values out of 100", len(values))
	}
}

func TestIntN(t *testing.T) {
	r := NewLockedRand()

	t.Run("produces values in range", func(t *testing.T) {
		for i := 0; i < 100; i++ {
			value := r.IntN(10)
			if value < 0 || value >= 10 {
				t.Fatalf("expected IntN(10) to return value in [0,10), got %d", value)
			}
		}
	})

	t.Run("produces different values", func(t *testing.T) {
		values := make(map[int]bool)
		for i := 0; i < 100; i++ {
			values[r.IntN(10)] = true
		}

		if len(values) < 3 {
			t.Fatalf("expected IntN(10) to produce different values, got only %d unique values out of 100", len(values))
		}
	})

	t.Run("boundary value", func(t *testing.T) {
		value := r.IntN(1)
		if value != 0 {
			t.Fatalf("expected IntN(1) to always return 0, got %d", value)
		}
	})
}

func TestUint32N(t *testing.T) {
	r := NewLockedRand()

	t.Run("produces values in range", func(t *testing.T) {
		for i := 0; i < 100; i++ {
			value := r.Uint32N(10)
			if value < 0 || value >= 10 {
				t.Fatalf("expected Uint32N(10) to return value in [0,10), got %d", value)
			}
		}
	})

	t.Run("produces different values", func(t *testing.T) {
		values := make(map[uint32]bool)
		for i := 0; i < 100; i++ {
			values[r.Uint32N(10)] = true
		}

		if len(values) < 3 {
			t.Fatalf("expected Uint32N(10) to produce different values, got only %d unique values out of 100", len(values))
		}
	})

	t.Run("boundary value", func(t *testing.T) {
		value := r.Uint32N(1)
		if value != 0 {
			t.Fatalf("expected Uint32N(1) to always return 0, got %d", value)
		}
	})
}

func TestShuffle(t *testing.T) {
	r := NewLockedRand()

	t.Run("shuffles array", func(t *testing.T) {
		original := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
		shuffled := make([]int, len(original))
		copy(shuffled, original)

		r.Shuffle(len(shuffled), func(i, j int) {
			shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
		})

		if reflect.DeepEqual(original, shuffled) {
			t.Fatalf("expected shuffled array to differ from original")
		}

		originalSum := 0
		shuffledSum := 0
		for i := range original {
			originalSum += original[i]
			shuffledSum += shuffled[i]
		}

		if originalSum != shuffledSum {
			t.Fatalf("expected sum to remain same after shuffle, got %d and %d", originalSum, shuffledSum)
		}
	})

	t.Run("handles empty array", func(t *testing.T) {
		empty := []int{}
		r.Shuffle(len(empty), func(i, j int) {})

		if len(empty) != 0 {
			t.Fatalf("expected empty array to remain empty")
		}
	})

	t.Run("handles single element", func(t *testing.T) {
		single := []int{42}
		r.Shuffle(len(single), func(i, j int) {
			single[i], single[j] = single[j], single[i]
		})

		if len(single) != 1 || single[0] != 42 {
			t.Fatalf("expected single element to remain unchanged")
		}
	})
}

func TestChaCha8Seed(t *testing.T) {
	seed1 := chaCha8Seed()
	seed2 := chaCha8Seed()

	if reflect.DeepEqual(seed1, seed2) {
		t.Fatalf("expected chaCha8Seed to produce different seeds on consecutive calls")
	}

	if len(seed1) != 32 {
		t.Fatalf("expected seed to be 32 bytes, got %d", len(seed1))
	}
}

func TestSeedFromUint64(t *testing.T) {
	input := uint64(12345)
	seed1 := seedFromUint64(input)
	seed2 := seedFromUint64(input)

	if !reflect.DeepEqual(seed1, seed2) {
		t.Fatalf("expected same input to produce same seed")
	}

	if len(seed1) != 32 {
		t.Fatalf("expected seed to be 32 bytes, got %d", len(seed1))
	}
}

func TestSplitMix64(t *testing.T) {
	value := uint64(12345)

	result1 := splitMix64(value)
	result2 := splitMix64(value)

	if result1 != result2 {
		t.Fatalf("expected splitMix64 to be deterministic")
	}

	if result1 == value {
		t.Fatalf("expected splitMix64 to produce different value from input")
	}
}

func TestSplitMix64Properties(t *testing.T) {
	value := uint64(0x9e3779b97f4a7c15)

	result := splitMix64(value)

	if result == 0 {
		t.Fatalf("expected splitMix64 to produce non-zero result for non-zero input")
	}

	if result == value {
		t.Fatalf("expected splitMix64 to produce different value from input")
	}
}

func TestConcurrentAccess(t *testing.T) {
	r := NewLockedRand()

	done := make(chan bool)
	results := make([]uint32, 100)

	for i := 0; i < 100; i++ {
		go func(idx int) {
			results[idx] = r.Uint32()
			done <- true
		}(i)
	}

	for i := 0; i < 100; i++ {
		<-done
	}

	for _, value := range results {
		if value > ^uint32(0) {
			t.Fatalf("expected all values to be valid uint32, got %d", value)
		}
	}
}

func TestDifferentSeedsProduceDifferentValues(t *testing.T) {
	seed1 := uint64(11111)
	seed2 := uint64(22222)

	r1 := NewLockedRandFromSeed(seed1)
	r2 := NewLockedRandFromSeed(seed2)

	value1 := r1.Uint32()
	value2 := r2.Uint32()

	if value1 == value2 {
		t.Fatalf("expected different seeds to produce different values, got %d for both", value1)
	}
}

func TestUint32Range(t *testing.T) {
	r := NewLockedRand()

	minValue := uint32(^uint32(0))
	maxValue := uint32(0)

	for i := 0; i < 1000; i++ {
		value := r.Uint32()
		if value < minValue {
			minValue = value
		}
		if value > maxValue {
			maxValue = value
		}
	}

	if minValue == maxValue {
		t.Fatalf("expected Uint32 to produce values across its range")
	}
}

func TestIntNZero(t *testing.T) {
	r := NewLockedRand()

	for i := 0; i < 10; i++ {
		value := r.IntN(1)
		if value != 0 {
			t.Fatalf("expected IntN(1) to always return 0, got %d", value)
		}
	}
}

func TestUint32NZero(t *testing.T) {
	r := NewLockedRand()

	for i := 0; i < 10; i++ {
		value := r.Uint32N(1)
		if value != 0 {
			t.Fatalf("expected Uint32N(1) to always return 0, got %d", value)
		}
	}
}
