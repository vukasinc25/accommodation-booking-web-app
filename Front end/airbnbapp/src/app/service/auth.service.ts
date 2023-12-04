import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Router } from '@angular/router';
import { JwtHelperService } from '@auth0/angular-jwt';
import { BehaviorSubject, Observable, Subject } from 'rxjs';

@Injectable({
  providedIn: 'root',
})
export class AuthService {
  private logginSubject: BehaviorSubject<boolean> =
    new BehaviorSubject<boolean>(false);

  private roleSubject: BehaviorSubject<string> = new BehaviorSubject<string>(
    ''
  );

  public isLoggedin: Observable<boolean>;

  public role: Observable<string>;

  constructor(private http: HttpClient, private router: Router) {
    this.isLoggedin = this.logginSubject.asObservable();
    this.role = this.roleSubject.asObservable();
  }

  private headers = new HttpHeaders({ 'Content-Type': 'application/json' });
  jwt: JwtHelperService = new JwtHelperService();

  login(user: any): Observable<any> {
    return this.http.post(
      '/api/users/login',
      { username: user.username, password: user.password },
      { headers: this.headers, responseType: 'json' }
    );
  }

  register(user: any): Observable<any> {
    return this.http.post(
      '/api/users/register',
      {
        username: user.username,
        password: user.password,
        role: user.userRole,
        email: user.email,
      },
      { headers: this.headers, responseType: 'json' }
    );
  }

  logout() {
    localStorage.removeItem('jwt');
    this.checkLoggin();
    this.checkRole();
  }

  checkLoggin(): boolean {
    if (!localStorage.getItem('jwt')) {
      this.setLoggin(false);
      return false;
    } else {
      this.setLoggin(true);
      return true;
    }
  }

  setLoggin(value: boolean) {
    this.logginSubject.next(value);
  }

  setRole(value: string) {
    this.roleSubject.next(value);
  }

  guardCheck(): string {
    let token = localStorage.getItem('jwt');
    if (token != null) return this.jwt.decodeToken(token).role;
    else return '';
  }

  checkRole() {
    let token = localStorage.getItem('jwt');
    if (token != null) this.setRole(this.jwt.decodeToken(token).role);
    else this.setRole('');
  }

  changeForgottenPassword(
    newPassword: string,
    confirmPassword: string,
    secretCode: string
  ): Observable<any> {
    return this.http.post(
      '/api/users/changeForgottenPassword',
      {
        newPassword: newPassword,
        confirmPassword: confirmPassword,
        code: secretCode,
      },
      { headers: this.headers, responseType: 'json' }
    );
  }

  sendForgottenPasswordEmail(email: string): Observable<any> {
    return this.http.post(
      `${'/api/users/sendforgottemail/'}${email}`,
      {},
      {
        headers: this.headers,
        responseType: 'json',
      }
    );
  }

  sendVerifyingEmail(code: string): Observable<any> {
    return this.http.post(
      `${'/api/users/email/'}${code}`,
      {},
      {
        headers: this.headers,
        responseType: 'json',
      }
    );
  }
}
