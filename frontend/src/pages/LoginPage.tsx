import { useState } from 'react';
import { useNavigate } from '@tanstack/react-router';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useTranslation } from 'react-i18next';
import { useAuthStore } from '@/stores/authStore';
import { api } from '@/api/client';

// Material Symbols icon component
function MaterialIcon({ name, className = '' }: { name: string; className?: string }) {
  return <span className={`material-symbols-outlined ${className}`}>{name}</span>;
}

const createLoginSchema = (t: (key: string) => string) => z.object({
  email: z.string().email(t('auth.validation.invalidEmail')),
  password: z.string().min(8, t('auth.validation.passwordMin')),
});

type LoginForm = z.infer<ReturnType<typeof createLoginSchema>>;

export function LoginPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { setAuth } = useAuthStore();
  const [error, setError] = useState<string | null>(null);
  const loginSchema = createLoginSchema(t);

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<LoginForm>({
    resolver: zodResolver(loginSchema),
  });

  const onSubmit = async (data: LoginForm) => {
    try {
      setError(null);
      const response = await api.auth.login(data);
      // ログインレスポンスからユーザー情報を取得
      setAuth(response.user, response.access_token, response.refresh_token);
      navigate({ to: '/' });
    } catch (err) {
      setError(err instanceof Error ? err.message : t('common.error'));
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-background">
      <div className="w-full max-w-md p-8 bg-card rounded-xl border border-border shadow-lg">
        {/* ロゴ */}
        <div className="text-center mb-8">
          <div className="inline-flex items-center justify-center size-16 bg-primary rounded-xl mb-4">
            <MaterialIcon name="schedule" className="text-4xl text-primary-foreground" />
          </div>
          <h1 className="text-2xl font-bold">{t('common.appName')}</h1>
          <p className="text-muted-foreground mt-2 text-sm">{t('auth.loginDescription')}</p>
        </div>

        <form onSubmit={handleSubmit(onSubmit)} className="space-y-5">
          {error && (
            <div className="flex items-center gap-3 p-4 text-sm text-destructive bg-destructive/10 rounded-lg border border-destructive/20">
              <MaterialIcon name="error" className="text-xl" />
              <span>{error}</span>
            </div>
          )}

          <div>
            <label className="block text-sm font-medium mb-2">{t('common.email')}</label>
            <div className="relative">
              <MaterialIcon name="mail" className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
              <input
                type="email"
                {...register('email')}
                className="w-full pl-10 pr-4 py-3 bg-black/20 border-none rounded-lg focus:outline-none focus:ring-2 focus:ring-primary text-sm"
                placeholder="user@example.com"
              />
            </div>
            {errors.email && (
              <p className="text-sm text-destructive mt-2 flex items-center gap-1">
                <MaterialIcon name="warning" className="text-sm" />
                {errors.email.message}
              </p>
            )}
          </div>

          <div>
            <label className="block text-sm font-medium mb-2">{t('common.password')}</label>
            <div className="relative">
              <MaterialIcon name="lock" className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
              <input
                type="password"
                {...register('password')}
                className="w-full pl-10 pr-4 py-3 bg-black/20 border-none rounded-lg focus:outline-none focus:ring-2 focus:ring-primary text-sm"
                placeholder="••••••••"
              />
            </div>
            {errors.password && (
              <p className="text-sm text-destructive mt-2 flex items-center gap-1">
                <MaterialIcon name="warning" className="text-sm" />
                {errors.password.message}
              </p>
            )}
          </div>

          <button
            type="submit"
            disabled={isSubmitting}
            className="w-full py-3 px-4 bg-primary text-primary-foreground font-bold rounded-lg hover:brightness-110 disabled:opacity-50 transition-all flex items-center justify-center gap-2"
          >
            {isSubmitting ? (
              <>
                <MaterialIcon name="progress_activity" className="animate-spin" />
                {t('common.loading')}
              </>
            ) : (
              <>
                <MaterialIcon name="login" />
                {t('auth.loginButton')}
              </>
            )}
          </button>
        </form>

        {/* Dev hint */}
        <div className="mt-6 p-4 bg-primary/5 rounded-lg border border-primary/20">
          <p className="text-xs text-muted-foreground text-center">
            {t('auth.devEnvironment')}: <code className="text-primary">admin@example.com</code> / <code className="text-primary">password123</code>
          </p>
        </div>
      </div>
    </div>
  );
}
