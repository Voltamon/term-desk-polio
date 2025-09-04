package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"
)

const (
	// Character set for access codes (alphanumeric)
	charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	codeLength = 9
)

func main() {
	fmt.Println("=== Access Code Generator ===")
	fmt.Printf("Generating %d-character alphanumeric access codes\n\n", codeLength)
	
	// Generate single access code
	code := GenerateAccessCode()
	fmt.Printf("Generated Access Code: %s\n\n", code)
	
	// Generate multiple codes
	fmt.Println("Generating 5 access codes:")
	for i := 0; i < 5; i++ {
		code := GenerateAccessCode()
		fmt.Printf("%d. %s\n", i+1, code)
	}
	
	// Generate codes with custom length
	fmt.Println("\nCustom length codes:")
	for _, length := range []int{6, 8, 12, 16} {
		code := GenerateCustomLengthCode(length)
		fmt.Printf("%d chars: %s\n", length, code)
	}
	
	// Generate formatted codes
	fmt.Println("\nFormatted codes (XXX-XXX-XXX):")
	for i := 0; i < 3; i++ {
		code := GenerateFormattedCode()
		fmt.Printf("%d. %s\n", i+1, code)
	}
	
	// Benchmark generation speed
	fmt.Println("\nBenchmarking...")
	benchmarkGeneration()
}

// GenerateAccessCode generates a 9-character alphanumeric access code
func GenerateAccessCode() string {
	return GenerateCustomLengthCode(codeLength)
}

// GenerateCustomLengthCode generates an alphanumeric code of specified length
func GenerateCustomLengthCode(length int) string {
	result := make([]byte, length)
	
	for i := range result {
		// Generate cryptographically secure random index
		index, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			log.Fatal("Error generating random number:", err)
		}
		result[i] = charset[index.Int64()]
	}
	
	return string(result)
}

// GenerateFormattedCode generates a formatted access code (XXX-XXX-XXX)
func GenerateFormattedCode() string {
	code := GenerateAccessCode()
	// Format as XXX-XXX-XXX
	return fmt.Sprintf("%s-%s-%s", code[:3], code[3:6], code[6:9])
}

// GenerateBatch generates multiple access codes
func GenerateBatch(count int) []string {
	codes := make([]string, count)
	for i := 0; i < count; i++ {
		codes[i] = GenerateAccessCode()
	}
	return codes
}

// GenerateUniqueCodeSet generates a set of unique access codes
func GenerateUniqueCodeSet(count int) []string {
	codeSet := make(map[string]bool)
	codes := make([]string, 0, count)
	
	for len(codes) < count {
		code := GenerateAccessCode()
		if !codeSet[code] {
			codeSet[code] = true
			codes = append(codes, code)
		}
	}
	
	return codes
}

// ValidateAccessCode validates if a code matches the expected format
func ValidateAccessCode(code string) bool {
	if len(code) != codeLength {
		return false
	}
	
	for _, char := range code {
		if !strings.ContainsRune(charset, char) {
			return false
		}
	}
	
	return true
}

// GenerateTimestampedCode generates a code with timestamp prefix
func GenerateTimestampedCode() string {
	timestamp := time.Now().Format("060102") // YYMMDD format
	code := GenerateCustomLengthCode(3)      // 3 random chars
	return timestamp + code                  // 6+3 = 9 chars total
}

// GenerateSecureCode generates a code using only uppercase letters and numbers
// excluding easily confused characters (0, O, I, 1)
func GenerateSecureCode() string {
	secureCharset := "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // Removed 0, O, I, 1
	result := make([]byte, codeLength)
	
	for i := range result {
		index, err := rand.Int(rand.Reader, big.NewInt(int64(len(secureCharset))))
		if err != nil {
			log.Fatal("Error generating random number:", err)
		}
		result[i] = secureCharset[index.Int64()]
	}
	
	return string(result)
}

// benchmarkGeneration tests the performance of code generation
func benchmarkGeneration() {
	start := time.Now()
	count := 10000
	
	for i := 0; i < count; i++ {
		GenerateAccessCode()
	}
	
	duration := time.Since(start)
	fmt.Printf("Generated %d codes in %v\n", count, duration)
	fmt.Printf("Average: %v per code\n", duration/time.Duration(count))
}

// Example usage functions

// GetUserAccessCode simulates getting an access code for a user
func GetUserAccessCode(userID string) string {
	code := GenerateAccessCode()
	fmt.Printf("Access code for user %s: %s\n", userID, code)
	return code
}

// GenerateSessionCode generates a session-specific access code
func GenerateSessionCode(sessionID string) string {
	code := GenerateAccessCode()
	fmt.Printf("Session %s access code: %s\n", sessionID, code)
	return code
}