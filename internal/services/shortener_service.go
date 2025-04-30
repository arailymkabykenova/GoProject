package services

import (
	"errors"
	"fmt"
	"log"
	"net/url"

	"template/internal/pkg/utils"
	"template/internal/repositories"
)

const (
	shortCodeLength      = 7
	maxGenerationRetries = 5
)

type ShortenerService interface {
	CreateShortURL(longURL string) (string, error)
	ValidateURL(inputURL string) bool
	UpdateLongURL(shortCode, newLongURL string) error
	DeleteMapping(shortCode string) error
}

type shortenerSvc struct {
	repo repositories.ShortenerRepository
}

func NewShortenerService(repo repositories.ShortenerRepository) ShortenerService {
	return &shortenerSvc{repo: repo}
}

func (s *shortenerSvc) CreateShortURL(longURL string) (string, error) {
	if !s.ValidateURL(longURL) {
		return "", errors.New("invalid URL format provided")
	}

	existingCode, err := s.repo.FindByLongURL(longURL)
	if err != nil && !errors.Is(err, repositories.ErrNotFound) {
		log.Printf("Service error checking for existing long URL '%s': %v", longURL, err)
		return "", fmt.Errorf("failed to check for existing URL: %w", err)
	}
	if existingCode != "" {
		log.Printf("Service found existing code '%s' for URL '%s'", existingCode, longURL)
		return existingCode, nil
	}

	for i := 0; i < maxGenerationRetries; i++ {
		code, err := utils.GenerateRandomString(shortCodeLength)
		if err != nil {
			return "", fmt.Errorf("service failed to generate random string: %w", err)
		}

		_, repoErr := s.repo.FindByShortCode(code)
		if repoErr != nil {
			if errors.Is(repoErr, repositories.ErrNotFound) {
				_, saveErr := s.repo.SaveMapping(code, longURL)
				if saveErr != nil {
					log.Printf("Service error saving new mapping (Code: %s): %v", code, saveErr)
					return "", fmt.Errorf("service failed to save mapping: %w", saveErr)
				}
				log.Printf("Service successfully created mapping: %s -> %s", code, longURL)
				return code, nil
			}
			log.Printf("Service database error checking code uniqueness (%s): %v", code, repoErr)
			return "", fmt.Errorf("service failed to check code uniqueness: %w", repoErr)
		}
		log.Printf("Service short code collision detected (%s), retrying (%d/%d)...", code, i+1, maxGenerationRetries)
	}

	log.Printf("Service failed to generate unique short code after %d retries", maxGenerationRetries)
	return "", fmt.Errorf("service could not generate unique short code after %d retries", maxGenerationRetries)
}

func (s *shortenerSvc) ValidateURL(inputURL string) bool {
	u, err := url.ParseRequestURI(inputURL)
	if err != nil {
		return false
	}
	return (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}

func (s *shortenerSvc) UpdateLongURL(shortCode, newLongURL string) error {
	if !s.ValidateURL(newLongURL) {
		return errors.New("invalid new URL format provided")
	}

	err := s.repo.UpdateLongURL(shortCode, newLongURL)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			log.Printf("Service: Attempted to update non-existent short code '%s'", shortCode)
			return err
		}
		log.Printf("Service error updating mapping for code '%s': %v", shortCode, err)
		return fmt.Errorf("service failed to update mapping: %w", err)
	}

	log.Printf("Service successfully updated mapping for code '%s' to '%s'", shortCode, newLongURL)
	return nil
}

func (s *shortenerSvc) DeleteMapping(shortCode string) error {
	err := s.repo.DeleteMapping(shortCode)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			log.Printf("Service: Attempted to delete non-existent short code '%s'", shortCode)
			return err
		}
		log.Printf("Service error deleting mapping for code '%s': %v", shortCode, err)
		return fmt.Errorf("service failed to delete mapping: %w", err)
	}

	log.Printf("Service successfully deleted mapping for code '%s'", shortCode)
	return nil
}
