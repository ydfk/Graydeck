package manager

import (
	"context"

	"mihomo-manager/internal/model"
)

func (s *Service) SyncSubscription(ctx context.Context, id string) (model.Subscription, error) {
	subscription, err := s.syncSubscription(ctx, id)
	if err != nil {
		return model.Subscription{}, err
	}

	if subscription.Enabled {
		if err := s.ensureRuntime(ctx); err != nil {
			s.appendLogf("应用当前配置失败：%s，%v", subscription.Name, err)
			return model.Subscription{}, err
		}
		s.appendLogf("配置文件更新完成并已应用：%s", subscription.Name)
		return s.findSubscription(id)
	}

	s.appendLogf("配置文件更新完成：%s", subscription.Name)
	return subscription, nil
}
