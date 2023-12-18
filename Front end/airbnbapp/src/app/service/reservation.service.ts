import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Router } from '@angular/router';
import { JwtHelperService } from '@auth0/angular-jwt';
import { NgbDate } from '@ng-bootstrap/ng-bootstrap';
import { Observable } from 'rxjs';

@Injectable({
  providedIn: 'root',
})
export class ReservationService {
  constructor(private http: HttpClient, private router: Router) {}

  private headers = new HttpHeaders({ 'Content-Type': 'application/json' });
  jwt: JwtHelperService = new JwtHelperService();

  createReservation(
    id: any,
    fromDate: NgbDate,
    toDate: NgbDate
  ): Observable<any> {
    let startDate = new Date(
      fromDate.year,
      fromDate.month - 1,
      fromDate.day + 1
    );
    let endDate = new Date(toDate.year, toDate.month - 1, toDate.day + 1);
    return this.http.post(
      '/api/reservations/date_for_acoo',
      {
        acco_id: id,
        begin_accomodation_date: startDate,
        end_accomodation_date: endDate,
      },
      { headers: this.headers, responseType: 'json' }
    );
  }

  getReservations(id: any): Observable<any> {
    return this.http.get('/api/reservations/dates_by_acco_id/' + id, {
      headers: this.headers,
      responseType: 'json',
    });
  }

  createReservationDatesForAccomodation(
    id: any,
    reservation: any
  ): Observable<any> {
    // console.log(reservation.pricePerPerson);
    // console.log(reservation.pricePerNight);
    // console.log(reservation.availableFrom);
    // console.log(reservation.availableUntil);
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
    // console.log(id);
    return this.http.get(`${'/api/reservations/by_acco/'}${id}`, {
      headers: this.headers,
      responseType: 'json',
    });
  }
}
