import { NgModule } from '@angular/core';
import { BrowserModule } from '@angular/platform-browser';

import { AppRoutingModule } from './app-routing.module';
import { AppComponent } from './app.component';
import { HeaderComponent } from './header/header.component';
import { MainPageComponent } from './main-page/main-page.component';
import { LoginComponent } from './login/login.component';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';
import { HTTP_INTERCEPTORS, HttpClientModule } from '@angular/common/http';
import { RegisterComponent } from './register/register.component';
import { AccommoAddComponent } from './accommo-add/accommo-add.component';
import { ProfileComponent } from './profile/profile.component';
import { TokenInterceptor } from './interceptor/token.interceptor';
import { NgxCaptchaModule } from 'ngx-captcha';
import { EmailVerificationPopupComponent } from './email-verification-popup/email-verification-popup.component';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
// import { MatDialogModule } from '@angular/material/dialog';
import { PasswordEmailRecoveryComponent } from './password-email-recovery/password-email-recovery.component';
import { ResetPasswordComponent } from './reset-password/reset-password.component';
import { VerifyEmailComponent } from './verify-email/verify-email.component';

@NgModule({
  declarations: [
    AppComponent,
    HeaderComponent,
    MainPageComponent,
    LoginComponent,
    RegisterComponent,
    AccommoAddComponent,
    ProfileComponent,
    EmailVerificationPopupComponent,
    PasswordEmailRecoveryComponent,
    ResetPasswordComponent,
    VerifyEmailComponent,
  ],
  imports: [
    BrowserModule,
    AppRoutingModule,
    FormsModule,
    ReactiveFormsModule,
    HttpClientModule,
    NgxCaptchaModule,
    BrowserAnimationsModule,
    // MatDialogModule,
  ],
  providers: [
    { provide: HTTP_INTERCEPTORS, useClass: TokenInterceptor, multi: true },
  ],
  bootstrap: [AppComponent],
})
export class AppModule {}
