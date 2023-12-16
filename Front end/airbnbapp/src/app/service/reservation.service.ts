import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Router } from '@angular/router';
import { JwtHelperService } from '@auth0/angular-jwt';
import { end } from '@popperjs/core';
import { Observable } from 'rxjs';

@Injectable({
  providedIn: 'root',
})
export class ReservationService {
  constructor(private http: HttpClient, private router: Router) {}

  private headers = new HttpHeaders({ 'Content-Type': 'application/json' });
  jwt: JwtHelperService = new JwtHelperService();

  createReservationDatesForAccomodation(
    id: any,
    reservation: any
  ): Observable<any> {
    console.log(reservation.pricePerPerson);
    console.log(reservation.pricePerNight);
    console.log(reservation.availableFrom);
    console.log(reservation.availableUntil);
    return this.http.post(
      '/api/reservations/for_acco',
      {
        accoId: id,
        numberPeople: 2,
        priceByPeople: reservation.pricePerPerson,
        priceByAccommodation: reservation.pricePerNight,
        startDate: reservation.availableFrom,
        endDate: reservation.availableUntil,
      },
      { headers: this.headers, responseType: 'json' }
    );
  }
  getAvailabelDatesForAccomodation(id: any): Observable<any> {
    console.log(id);
    return this.http.get(`${'/api/reservations/by_acco/'}${id}`, {
      headers: this.headers,
      responseType: 'json',
    });
  }
}
