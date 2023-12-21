import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { MainPageComponent } from './main-page/main-page.component';
import { LoginComponent } from './login/login.component';
import { RegisterComponent } from './register/register.component';
import { AccommoAddComponent } from './accommo-add/accommo-add.component';
import { ResetPasswordComponent } from './reset-password/reset-password.component';
import { ProfileComponent } from './profile/profile.component';
import { roleGuard } from './guard/role.guard';
import { loginGuard } from './guard/login.guard';
import { PasswordEmailRecoveryComponent } from './password-email-recovery/password-email-recovery.component';
import { AccommoInfoComponent } from './accommo-info/accommo-info.component';
import { VerifyEmailComponent } from './verify-email/verify-email.component';
import { MyAccommoComponent } from './my-accommo/my-accommo.component';

const routes: Routes = [
  {
    path: '',
    component: MainPageComponent,
  },
  {
    path: 'sendEmail',
    component: PasswordEmailRecoveryComponent,
  },
  {
    path: 'login',
    component: LoginComponent,
  },
  {
    path: 'register',
    component: RegisterComponent,
  },
  {
    path: 'profile',
    component: ProfileComponent,
    canActivate: [loginGuard],
  },
  {
    path: 'accommodations/info/:id',
    component: AccommoInfoComponent,
  },
  {
    path: 'accommodations/myAccommodations',
    component: MyAccommoComponent,
  },
  {
    path: 'accommodations/create',
    component: AccommoAddComponent,
    canActivate: [loginGuard, roleGuard],
  },
  {
    path: 'verify-email',
    component: VerifyEmailComponent,
  },
  { path: 'reset-password', component: ResetPasswordComponent },
  {
    path: '**',
    redirectTo: '',
    pathMatch: 'full',
  },
];

@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule],
})
export class AppRoutingModule {}
