import { Component } from '@angular/core';
import { FormGroup, FormBuilder, Validators } from '@angular/forms';
import { Router } from '@angular/router';
import { AuthService } from '../service/auth.service';
// import { MatDialog } from '@angular/material/dialog';
import { EmailVerificationPopupComponent } from '../email-verification-popup/email-verification-popup.component';

@Component({
  selector: 'app-register',
  templateUrl: './register.component.html',
  styleUrls: ['./register.component.css'],
})
export class RegisterComponent {
  form: FormGroup;
  siteKey: string;

  constructor(
    // private dialogRef: MatDialog,
    private fb: FormBuilder,
    private router: Router,
    private authService: AuthService
  ) {
    this.form = this.fb.group({
      username: [null, Validators.required],
      email: ['', Validators.compose([Validators.required, Validators.email])],
      firstName: [null, Validators.required],
      lastName: [null, Validators.required],
      country: [null, Validators.required],
      city: [null, Validators.required],
      streetName: [null, Validators.required],
      streetNumber: [null, Validators.required],
      password: [null, Validators.required],
      userRole: [null, Validators.required],
    });

    this.siteKey = '6LddmB4pAAAAALdViM1b2M9OJZNgwKQ-HbFtGXK-';
  }

  submit() {
    this.authService.register(this.form.value).subscribe({
      next: (data) => {
        console.log('Register success');
        this.router.navigate(['login']);
      },
      error: (err) => {
        console.log(err);
        alert(err.error.message);
      },
    });
  }

  get email() {
    return this.form.get('email');
  }

  get password() {
    return this.form.get('password');
  }

  // openDialog() {
  //   this.dialogRef.open(EmailVerificationPopupComponent);
  // }
}
