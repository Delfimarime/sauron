package usecase

import (
	"context"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/delfimarime/sauron/internal/infrastructure/repository/storage"
	"github.com/delfimarime/sauron/internal/telemetry"
	"github.com/delfimarime/sauron/pkg/sauron/types"
)

// DescribeProviderUseCaseParams injects the collaborators the use case composes.
type DescribeProviderUseCaseParams struct {
	fx.In
	Providers storage.ProvidersStore
	Logger    *zap.Logger
}

// DescribeProviderUseCase reads the configured provider and returns its detail;
// field selection and directory derivation are presentation concerns resolved by
// the caller.
type DescribeProviderUseCase struct {
	logger    *zap.Logger
	providers storage.ProvidersStore
}

// NewDescribeProviderUseCase builds the use case from the injected collaborators.
func NewDescribeProviderUseCase(params DescribeProviderUseCaseParams) *DescribeProviderUseCase {
	return &DescribeProviderUseCase{
		providers: params.Providers,
		logger:    params.Logger,
	}
}

// Execute reads the configured provider, classifying a read failure as an io
// error. When none is set it returns (nil, nil) — that is not an error: the
// caller reports none-set and exits successfully. Otherwise it returns the
// configured provider.
func (uc *DescribeProviderUseCase) Execute(ctx context.Context, _ DescribeProviderRequest) (*types.Provider, error) {
	provider, err := uc.providers.Get(ctx)
	if err != nil {
		return nil, ioErr("read provider", err)
	}
	if provider == nil {
		return nil, nil
	}

	uc.logger.Debug("provider described",
		zap.String(telemetry.FieldProviderName, provider.Metadata.Name),
	)

	return provider, nil
}
