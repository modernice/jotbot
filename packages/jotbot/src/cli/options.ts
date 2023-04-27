export interface WithVerboseOption {
  verbose?: boolean
}

export interface WithSourceOption {
  path?: string
}

export const verboseOption = ['-v, --verbose', 'Verbose output'] as const
