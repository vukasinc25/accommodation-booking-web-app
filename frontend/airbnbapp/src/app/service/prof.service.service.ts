import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Router } from '@angular/router';
import { JwtHelperService } from '@auth0/angular-jwt';
import { Observable } from 'rxjs';

@Injectable({
  providedIn: 'root',
})
export class ProfServiceService {
  constructor(private http: HttpClient, private router: Router) {}

  private headers = new HttpHeaders({ 'Content-Type': 'application/json' });
  jwt: JwtHelperService = new JwtHelperService();

  getUserInfo(): Observable<any> {
    return this.http.get('/api/prof/user', {
      headers: this.headers,
      responseType: 'json',
    });
  }

  updateUserInfo(user: any): Observable<any> {
    return this.http.patch(
      '/api/users/update',
      {
        email: user.email,
        firstname: user.firstName,
        lastname: user.lastName,
        location: {
          country: user.country,
          city: user.city,
          streetName: user.streetName,
          streetNumber: user.streetNumber,
        },
      },
      {
        headers: this.headers,
        responseType: 'json',
      }
    );
  }
}
