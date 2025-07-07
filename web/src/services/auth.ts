import api from './api';

// 认证相关的类型定义
export interface LoginRequest {
  username: string;
  password: string;
}

export interface LoginResponse {
  access_token: string;
  refresh_token: string;
  token_type: string;
  expires_in: number;
  user: UserInfo;
}

export interface RegisterRequest {
  username: string;
  email: string;
  password: string;
  full_name?: string;
}

export interface RegisterResponse {
  id: number;
  username: string;
  email: string;
  full_name?: string;
  message: string;
  created_at: string;
}

export interface RefreshTokenRequest {
  refresh_token: string;
}

export interface RefreshTokenResponse {
  access_token: string;
  refresh_token: string;
  token_type: string;
  expires_in: number;
}

export interface UserInfo {
  id: number;
  username: string;
  email: string;
  full_name?: string;
  balance?: number;
}

export interface UserProfile {
  id: number;
  username: string;
  email: string;
  full_name?: string;
  balance: number;
  created_at: string;
  updated_at: string;
}

export interface ChangePasswordRequest {
  old_password: string;
  new_password: string;
}

export interface RechargeRequest {
  amount: number;
  description?: string;
}

// 认证服务类
export class AuthService {
  /**
   * 用户登录 - 不需要token认证
   */
  static async login(credentials: LoginRequest): Promise<LoginResponse> {
    try {
      // 使用noAuth方法，确保不会注入token
      const response = await api.noAuth.post<LoginResponse>('/auth/login', credentials);

      if (response.success && response.data) {
        // 存储token到localStorage
        localStorage.setItem('access_token', response.data.access_token);
        localStorage.setItem('refresh_token', response.data.refresh_token);
        localStorage.setItem('user_info', JSON.stringify(response.data.user));

        return response.data;
      }

      throw new Error(response.error?.message || 'Login failed');
    } catch (error: any) {
      // 处理401错误（用户名或密码错误）
      if (error.response?.status === 401) {
        throw new Error('INVALID_CREDENTIALS');
      }

      // 处理其他错误
      if (error.response?.data?.error?.message) {
        throw new Error(error.response.data.error.message);
      }

      throw new Error(error.message || 'Login failed');
    }
  }

  /**
   * 用户注册 - 不需要token认证
   */
  static async register(userData: RegisterRequest): Promise<RegisterResponse> {
    // 使用noAuth方法，确保不会注入token
    const response = await api.noAuth.post<RegisterResponse>('/auth/register', userData);

    if (response.success && response.data) {
      return response.data;
    }

    throw new Error(response.error?.message || 'Registration failed');
  }

  /**
   * 刷新访问令牌 - 不需要token认证
   */
  static async refreshToken(refreshToken: string): Promise<RefreshTokenResponse> {
    // 使用noAuth方法，确保不会注入token
    const response = await api.noAuth.post<RefreshTokenResponse>('/auth/refresh', {
      refresh_token: refreshToken,
    });

    if (response.success && response.data) {
      // 更新存储的token
      localStorage.setItem('access_token', response.data.access_token);
      localStorage.setItem('refresh_token', response.data.refresh_token);

      return response.data;
    }

    throw new Error(response.error?.message || 'Token refresh failed');
  }

  /**
   * 获取用户资料
   */
  static async getProfile(): Promise<UserProfile> {
    const response = await api.get<UserProfile>('/auth/profile');

    if (response.success && response.data) {
      return response.data;
    }

    throw new Error(response.error?.message || 'Failed to get user profile');
  }

  /**
   * 修改密码
   */
  static async changePassword(passwordData: ChangePasswordRequest): Promise<void> {
    const response = await api.post('/auth/change-password', passwordData);

    if (!response.success) {
      throw new Error(response.error?.message || 'Password change failed');
    }
  }

  /**
   * 充值余额
   */
  static async recharge(rechargeData: RechargeRequest): Promise<UserProfile> {
    const response = await api.post<UserProfile>('/auth/recharge', {
      amount: rechargeData.amount,
      description: rechargeData.description || '用户充值'
    });

    if (response.success && response.data) {
      return response.data;
    }

    throw new Error(response.error?.message || 'Failed to recharge');
  }

  /**
   * 用户登出
   */
  static logout(): void {
    // 清除本地存储的认证信息
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
    localStorage.removeItem('user_info');
    
    // 跳转到登录页
    window.location.href = '/sign-in';
  }

  /**
   * 检查用户是否已登录
   */
  static isAuthenticated(): boolean {
    const token = localStorage.getItem('access_token');
    return !!token;
  }

  /**
   * 获取当前用户信息
   */
  static getCurrentUser(): UserInfo | null {
    const userInfo = localStorage.getItem('user_info');
    if (userInfo) {
      try {
        return JSON.parse(userInfo);
      } catch (error) {
        console.error('Failed to parse user info:', error);
        return null;
      }
    }
    return null;
  }

  /**
   * 获取访问令牌
   */
  static getAccessToken(): string | null {
    return localStorage.getItem('access_token');
  }

  /**
   * 获取刷新令牌
   */
  static getRefreshToken(): string | null {
    return localStorage.getItem('refresh_token');
  }

  /**
   * 检查token是否即将过期（提前5分钟刷新）
   */
  static shouldRefreshToken(): boolean {
    const token = this.getAccessToken();
    if (!token) return false;

    try {
      // 解析JWT token的payload
      const payload = JSON.parse(atob(token.split('.')[1]));
      const expirationTime = payload.exp * 1000; // 转换为毫秒
      const currentTime = Date.now();
      const fiveMinutes = 5 * 60 * 1000; // 5分钟

      // 如果token在5分钟内过期，则需要刷新
      return expirationTime - currentTime < fiveMinutes;
    } catch (error) {
      console.error('Failed to parse token:', error);
      return true; // 解析失败时也尝试刷新
    }
  }

  /**
   * 自动刷新token（如果需要）
   */
  static async autoRefreshToken(): Promise<void> {
    if (this.shouldRefreshToken()) {
      const refreshToken = this.getRefreshToken();
      if (refreshToken) {
        try {
          await this.refreshToken(refreshToken);
        } catch (error) {
          console.error('Auto refresh token failed:', error);
          this.logout();
        }
      }
    }
  }
}

export default AuthService;
