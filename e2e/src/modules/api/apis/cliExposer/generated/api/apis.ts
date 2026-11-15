export * from './healthApi';
import { HealthApi } from './healthApi';
export * from './researcherCommandLineInterfaceApi';
import { ResearcherCommandLineInterfaceApi } from './researcherCommandLineInterfaceApi';
export * from './researcherCommandLineInterfaceDeprecatedApi';
import { ResearcherCommandLineInterfaceDeprecatedApi } from './researcherCommandLineInterfaceDeprecatedApi';
import { HttpError } from './httpError';
export { HttpError };

export type { RequestFile } from '../model/models';

export const APIS = [HealthApi, ResearcherCommandLineInterfaceApi, ResearcherCommandLineInterfaceDeprecatedApi];


