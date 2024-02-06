import { CanActivateFn, Router } from '@angular/router';
import { AuthService } from '../../service/auth.service';
import { inject } from '@angular/core';
import { map } from 'rxjs';

export const loginGuard: CanActivateFn = (route, state) => {
  const authService = inject(AuthService);
  const router = inject(Router);

  if (authService.checkLoggin()) {
    return true;
  } else {
    router.navigate(['login']);
    return false;
  }
};
