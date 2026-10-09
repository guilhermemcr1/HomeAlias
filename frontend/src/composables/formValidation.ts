import type { InjectionKey, Ref } from 'vue'

export const formValidationCycle: InjectionKey<Ref<number>> = Symbol('formValidationCycle')
