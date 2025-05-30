import { DataService } from './dataService.interface';
import { LocalDataService } from './localDataService';
import { CloudDataService } from './cloudDataService';
import { store } from '../../store';

class DataServiceFactory {
  private localService: LocalDataService;
  private cloudService: CloudDataService;

  constructor() {
    this.localService = new LocalDataService();
    this.cloudService = new CloudDataService();
  }

  getCurrentService(): DataService {
    const state = store.getState();
    const isAuthenticated = state.auth.isAuthenticated;
    
    return isAuthenticated ? this.cloudService : this.localService;
  }

  getLocalService(): LocalDataService {
    return this.localService;
  }

  getCloudService(): CloudDataService {
    return this.cloudService;
  }
}

export const dataServiceFactory = new DataServiceFactory();
export const getDataService = () => dataServiceFactory.getCurrentService();