package service

import (
	"crm-project/internal/models/entity"
	"crm-project/internal/repository/postgres"
	"sync"
)

type SearchService struct {
	LeadRepository *postgres.LeadRepository
	DealRepository *postgres.DealRepository
	UserRepository *postgres.UserRepository
}

func NewSearchService(leadRepo *postgres.LeadRepository, dealRepo *postgres.DealRepository, userRepo *postgres.UserRepository) *SearchService {
	return &SearchService{LeadRepository: leadRepo, DealRepository: dealRepo, UserRepository: userRepo}
}

func (r *SearchService) GlobalSearch(keyword string) (entity.GlobalSearchResponse, error) {
	var (
		wg    sync.WaitGroup
		mu    sync.Mutex
		leads []entity.Lead
		deals []entity.Deal
		users []entity.User
		errs  []error
	)

	wg.Add(3)
	go func() {
		defer wg.Done()
		result, err := r.LeadRepository.SearchLeads(keyword)
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			errs = append(errs, err)
			return
		}
		leads = result
	}()

	go func() {
		defer wg.Done()
		result, err := r.DealRepository.SearchDeals(keyword)
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			errs = append(errs, err)
			return
		}
		deals = result
	}()

	go func() {
		defer wg.Done()
		result, err := r.UserRepository.SearchUsers(keyword)
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			errs = append(errs, err)
			return
		}
		users = result
	}()

	wg.Wait()
	if len(errs) > 0 {
		return entity.
			GlobalSearchResponse{}, errs[0]
	}

	result := entity.GlobalSearchResponse{
		Leads: leads,
		Deals: deals,
		Users: users,
	}
	return result, nil
}
