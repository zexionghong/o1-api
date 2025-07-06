import { useState, useCallback, useMemo } from 'react';
import { useTranslation } from 'react-i18next';

// 验证规则类型
export interface ValidationRule {
  required?: boolean;
  minLength?: number;
  maxLength?: number;
  pattern?: RegExp;
  email?: boolean;
  password?: boolean;
  confirmPassword?: string; // 字段名，用于密码确认
  custom?: (value: any, formData: any) => string | null;
  message?: string; // 自定义错误消息
}

// 字段配置类型
export interface FieldConfig {
  [key: string]: ValidationRule;
}

// 表单状态类型
export interface FormState<T> {
  values: T;
  errors: Partial<Record<keyof T, string>>;
  touched: Partial<Record<keyof T, boolean>>;
  isValid: boolean;
  isSubmitting: boolean;
  isDirty: boolean;
}

// 表单选项类型
export interface FormOptions<T> {
  initialValues: T;
  validationRules?: FieldConfig;
  onSubmit?: (values: T) => Promise<void> | void;
  validateOnChange?: boolean;
  validateOnBlur?: boolean;
}

// 预定义验证规则
export const validationRules = {
  required: (message?: string): ValidationRule => ({
    required: true,
    message,
  }),

  email: (message?: string): ValidationRule => ({
    email: true,
    pattern: /^[^\s@]+@[^\s@]+\.[^\s@]+$/,
    message,
  }),

  password: (minLength = 6, message?: string): ValidationRule => ({
    password: true,
    minLength,
    pattern: /^(?=.*[a-zA-Z])(?=.*\d)/,
    message,
  }),

  minLength: (length: number, message?: string): ValidationRule => ({
    minLength: length,
    message,
  }),

  maxLength: (length: number, message?: string): ValidationRule => ({
    maxLength: length,
    message,
  }),

  pattern: (regex: RegExp, message?: string): ValidationRule => ({
    pattern: regex,
    message,
  }),

  confirmPassword: (passwordField: string, message?: string): ValidationRule => ({
    confirmPassword: passwordField,
    message,
  }),

  custom: (validator: (value: any, formData: any) => string | null, message?: string): ValidationRule => ({
    custom: validator,
    message,
  }),
};

// 通用表单Hook
export function useForm<T extends Record<string, any>>(options: FormOptions<T>) {
  const { t } = useTranslation();
  const {
    initialValues,
    validationRules: rules = {},
    onSubmit,
    validateOnChange = true,
    validateOnBlur = true,
  } = options;

  const [state, setState] = useState<FormState<T>>({
    values: { ...initialValues },
    errors: {},
    touched: {},
    isValid: true,
    isSubmitting: false,
    isDirty: false,
  });

  // 验证单个字段
  const validateField = useCallback((name: keyof T, value: any, allValues: T): string | null => {
    const rule = rules[name as string];
    if (!rule) return null;

    // 必填验证
    if (rule.required && (!value || (typeof value === 'string' && !value.trim()))) {
      return rule.message || t('validation.required');
    }

    // 如果值为空且不是必填，跳过其他验证
    if (!value || (typeof value === 'string' && !value.trim())) {
      return null;
    }

    // 最小长度验证
    if (rule.minLength && value.length < rule.minLength) {
      return rule.message || t('validation.min_length', { length: rule.minLength });
    }

    // 最大长度验证
    if (rule.maxLength && value.length > rule.maxLength) {
      return rule.message || t('validation.max_length', { length: rule.maxLength });
    }

    // 正则表达式验证
    if (rule.pattern && !rule.pattern.test(value)) {
      if (rule.email) {
        return rule.message || t('validation.invalid_email');
      }
      if (rule.password) {
        return rule.message || t('validation.invalid_password');
      }
      return rule.message || t('validation.invalid_format');
    }

    // 密码确认验证
    if (rule.confirmPassword) {
      const passwordValue = allValues[rule.confirmPassword as keyof T];
      if (value !== passwordValue) {
        return rule.message || t('validation.passwords_not_match');
      }
    }

    // 自定义验证
    if (rule.custom) {
      return rule.custom(value, allValues);
    }

    return null;
  }, [rules, t]);

  // 验证所有字段
  const validateForm = useCallback((values: T): Partial<Record<keyof T, string>> => {
    const errors: Partial<Record<keyof T, string>> = {};
    
    Object.keys(rules).forEach((fieldName) => {
      const error = validateField(fieldName as keyof T, values[fieldName as keyof T], values);
      if (error) {
        errors[fieldName as keyof T] = error;
      }
    });

    return errors;
  }, [rules, validateField]);

  // 设置字段值
  const setValue = useCallback((name: keyof T, value: any) => {
    setState(prev => {
      const newValues = { ...prev.values, [name]: value };
      const newErrors = { ...prev.errors };
      
      // 验证当前字段
      if (validateOnChange) {
        const error = validateField(name, value, newValues);
        if (error) {
          newErrors[name] = error;
        } else {
          delete newErrors[name];
        }
      }

      const isValid = Object.keys(newErrors).length === 0;
      const isDirty = JSON.stringify(newValues) !== JSON.stringify(initialValues);

      return {
        ...prev,
        values: newValues,
        errors: newErrors,
        isValid,
        isDirty,
      };
    });
  }, [validateField, validateOnChange, initialValues]);

  // 设置字段错误
  const setError = useCallback((name: keyof T, error: string) => {
    setState(prev => ({
      ...prev,
      errors: { ...prev.errors, [name]: error },
      isValid: false,
    }));
  }, []);

  // 清除字段错误
  const clearError = useCallback((name: keyof T) => {
    setState(prev => {
      const newErrors = { ...prev.errors };
      delete newErrors[name];
      return {
        ...prev,
        errors: newErrors,
        isValid: Object.keys(newErrors).length === 0,
      };
    });
  }, []);

  // 设置字段为已触摸
  const setTouched = useCallback((name: keyof T, touched = true) => {
    setState(prev => ({
      ...prev,
      touched: { ...prev.touched, [name]: touched },
    }));
  }, []);

  // 处理字段变化
  const handleChange = useCallback((name: keyof T) => (
    event: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement> | any
  ) => {
    const value = event?.target?.value ?? event;
    setValue(name, value);
  }, [setValue]);

  // 处理字段失焦
  const handleBlur = useCallback((name: keyof T) => () => {
    setTouched(name, true);
    
    if (validateOnBlur) {
      const error = validateField(name, state.values[name], state.values);
      if (error) {
        setError(name, error);
      } else {
        clearError(name);
      }
    }
  }, [state.values, validateField, validateOnBlur, setTouched, setError, clearError]);

  // 提交表单
  const handleSubmit = useCallback(async (event?: React.FormEvent) => {
    if (event) {
      event.preventDefault();
    }

    // 验证所有字段
    const errors = validateForm(state.values);
    const isValid = Object.keys(errors).length === 0;

    // 标记所有字段为已触摸
    const allTouched = Object.keys(state.values).reduce((acc, key) => {
      acc[key as keyof T] = true;
      return acc;
    }, {} as Partial<Record<keyof T, boolean>>);

    setState(prev => ({
      ...prev,
      errors,
      touched: allTouched,
      isValid,
      isSubmitting: isValid,
    }));

    if (isValid && onSubmit) {
      try {
        await onSubmit(state.values);
      } catch (error) {
        console.error('Form submission error:', error);
      } finally {
        setState(prev => ({ ...prev, isSubmitting: false }));
      }
    }

    return isValid;
  }, [state.values, validateForm, onSubmit]);

  // 重置表单
  const reset = useCallback(() => {
    setState({
      values: { ...initialValues },
      errors: {},
      touched: {},
      isValid: true,
      isSubmitting: false,
      isDirty: false,
    });
  }, [initialValues]);

  // 获取字段属性
  const getFieldProps = useCallback((name: keyof T) => ({
    value: state.values[name] || '',
    onChange: handleChange(name),
    onBlur: handleBlur(name),
    error: state.touched[name] && !!state.errors[name],
    helperText: state.touched[name] ? state.errors[name] : '',
  }), [state, handleChange, handleBlur]);

  return {
    values: state.values,
    errors: state.errors,
    touched: state.touched,
    isValid: state.isValid,
    isSubmitting: state.isSubmitting,
    isDirty: state.isDirty,
    setValue,
    setError,
    clearError,
    setTouched,
    handleChange,
    handleBlur,
    handleSubmit,
    reset,
    getFieldProps,
    validateForm: () => validateForm(state.values),
  };
}
