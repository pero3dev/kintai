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
    <div className="min-h-screen flex items-center justify-center relative overflow-hidden">
      {/* Aurora Background */}
      <div className="aurora-bg">
        <div className="aurora-orb-1" />
        <div className="aurora-orb-2" />
      </div>

      {/* Noise Overlay */}
      <div className="noise-overlay" />

      {/* Login Card */}
      <div className="w-full max-w-md p-8 glass-card rounded-2xl relative z-10 animate-scale-in">
        {/* ロゴ */}
        <div className="text-center mb-8">
          <div className="inline-flex items-center justify-center size-16 gradient-primary rounded-2xl mb-4 shadow-glow-md">
            <MaterialIcon name="schedule" className="text-4xl text-white" />
          </div>
          <h1 className="text-2xl font-bold gradient-text">{t('common.appName')}</h1>
          <p className="text-muted-foreground mt-2 text-sm">{t('auth.loginDescription')}</p>
        </div>

        <form onSubmit={handleSubmit(onSubmit)} className="space-y-5">
          {error && (
            <div className="flex items-center gap-3 p-4 text-sm text-red-400 bg-red-500/10 rounded-xl border border-red-500/20">
              <MaterialIcon name="error" className="text-xl" />
              <span>{error}</span>
            </div>
          )}

          <div>
            <label className="block text-sm font-medium mb-2 text-foreground/80">{t('common.email')}</label>
            <div className="relative">
              <MaterialIcon name="mail" className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
              <input
                type="email"
                {...register('email')}
                className="w-full pl-10 pr-4 py-3 glass-input rounded-xl text-sm text-foreground placeholder:text-muted-foreground"
                placeholder="user@example.com"
              />
            </div>
            {errors.email && (
              <p className="text-sm text-red-400 mt-2 flex items-center gap-1">
                <MaterialIcon name="warning" className="text-sm" />
                {errors.email.message}
              </p>
            )}
          </div>

          <div>
            <label className="block text-sm font-medium mb-2 text-foreground/80">{t('common.password')}</label>
            <div className="relative">
              <MaterialIcon name="lock" className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
              <input
                type="password"
                {...register('password')}
                className="w-full pl-10 pr-4 py-3 glass-input rounded-xl text-sm text-foreground placeholder:text-muted-foreground"
                placeholder="••••••••"
              />
            </div>
            {errors.password && (
              <p className="text-sm text-red-400 mt-2 flex items-center gap-1">
                <MaterialIcon name="warning" className="text-sm" />
                {errors.password.message}
              </p>
            )}
          </div>

          <button
            type="submit"
            disabled={isSubmitting}
            className="w-full py-3 px-4 gradient-primary text-white font-semibold rounded-xl hover:shadow-glow-md disabled:opacity-50 transition-all duration-300 flex items-center justify-center gap-2"
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
        <div className="mt-6 p-4 glass-subtle rounded-xl">
          <p className="text-xs text-muted-foreground text-center">
            {t('auth.devEnvironment')}: <code className="text-indigo-400">admin@example.com</code> / <code className="text-indigo-400">password123</code>
          </p>
        </div>
      </div>
    </div>
  );
}
