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
import { AccommoInfoComponent } from './accommo-info/accommo-info.component';
import { EmailVerificationPopupComponent } from './email-verification-popup/email-verification-popup.component';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
// import { MatDialogModule } from '@angular/material/dialog';
import { PasswordEmailRecoveryComponent } from './password-email-recovery/password-email-recovery.component';
import { ResetPasswordComponent } from './reset-password/reset-password.component';
import { VerifyEmailComponent } from './verify-email/verify-email.component';
import { NgbModule } from '@ng-bootstrap/ng-bootstrap';
import { MyAccommoComponent } from './my-accommo/my-accommo.component';
import { AccommoListComponent } from './accommo-list/accommo-list.component';

//FA
import { FontAwesomeModule } from '@fortawesome/angular-fontawesome';
import { faMagnifyingGlass } from '@fortawesome/free-solid-svg-icons';

@NgModule({
  declarations: [
    AppComponent,
    HeaderComponent,
    MainPageComponent,
    LoginComponent,
    RegisterComponent,
    AccommoAddComponent,
    ProfileComponent,
    AccommoInfoComponent,
    EmailVerificationPopupComponent,
    PasswordEmailRecoveryComponent,
    ResetPasswordComponent,
    VerifyEmailComponent,
    MyAccommoComponent,
    AccommoListComponent,
  ],
  imports: [
    BrowserModule,
    AppRoutingModule,
    FormsModule,
    ReactiveFormsModule,
    HttpClientModule,
    NgxCaptchaModule,
    BrowserAnimationsModule,
    NgbModule,
    FontAwesomeModule,
    // MatDialogModule,
  ],
  providers: [
    { provide: HTTP_INTERCEPTORS, useClass: TokenInterceptor, multi: true },
  ],
  bootstrap: [AppComponent],
})
export class AppModule {
  faMagnifyingGlass = faMagnifyingGlass;
}
